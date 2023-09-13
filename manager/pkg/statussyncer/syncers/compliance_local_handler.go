package dbsyncer

import (
	"context"
	"fmt"
	"sync"

	"github.com/cenkalti/backoff/v4"
	set "github.com/deckarep/golang-set"
	"gorm.io/gorm"

	"github.com/stolostron/multicluster-global-hub/pkg/bundle/status"
	"github.com/stolostron/multicluster-global-hub/pkg/database"
	"github.com/stolostron/multicluster-global-hub/pkg/database/models"
)

func (syncer *PoliciesDBSyncer) handleLocalClustersPerPolicyBundle(ctx context.Context, bundle status.Bundle) error {
	logBundleHandlingMessage(syncer.log, bundle, startBundleHandlingMessage)
	leafHubName := bundle.GetLeafHubName()
	db := database.GetGorm()

	// policyID: { compliance: (cluster1, cluster2), nonCompliance: (cluster3, cluster4), unknowns: (cluster5) }
	allPolicyClusterSetsFromDB, err := getAllLocalPolicyClusterSets(db, "leaf_hub_name = ?", leafHubName)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup // Create a WaitGroup

	for _, object := range bundle.GetObjects() { // every object is clusters list per policy with full state
		clustersPerPolicyFromBundle, ok := object.(*status.PolicyGenericComplianceStatus)
		if !ok {
			continue // do not handle objects other than PolicyGenericComplianceStatus
		}

		policyClusterSetFromDB, policyExistsInDB :=
			allPolicyClusterSetsFromDB[clustersPerPolicyFromBundle.PolicyID]
		if !policyExistsInDB {
			policyClusterSetFromDB = NewPolicyClusterSets()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			err = backoff.Retry(func() error {
				return db.Transaction(func(tx *gorm.DB) error {
					e := handleLocalComplianceWithTransaction(db, leafHubName,
						policyClusterSetFromDB, clustersPerPolicyFromBundle)
					return e
				})
			}, backoff.NewExponentialBackOff())

			if err != nil {
				syncer.log.Error(err, "failed to handle clusters per policy bundle", "policyID",
					clustersPerPolicyFromBundle.PolicyID)
			} else {
				// keep this policy in db, should remove from db only policies that were not sent in the bundle
				delete(allPolicyClusterSetsFromDB, clustersPerPolicyFromBundle.PolicyID)
			}
		}()
	}

	wg.Wait()

	// remove policies that were not sent in the bundle
	err = backoff.Retry(func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			for policyID := range allPolicyClusterSetsFromDB {
				err := tx.Where(&models.LocalStatusCompliance{
					PolicyID: policyID,
				}).Delete(&models.LocalStatusCompliance{}).Error
				if err != nil {
					return err
				}
			}
			return nil
		})
	}, backoff.NewExponentialBackOff())

	if err != nil {
		return fmt.Errorf("failed to delete clusters per policy bundle - %w", err)
	}
	logBundleHandlingMessage(syncer.log, bundle, finishBundleHandlingMessage)
	return nil
}

func handleLocalComplianceWithTransaction(db *gorm.DB, leafHubName string, clusterSetFromDB *PolicyClustersSets,
	clusterPerPolicyBundle *status.PolicyGenericComplianceStatus,
) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		// handle compliant clusters of the policy
		allClustersOnDB := clusterSetFromDB.GetAllClusters()
		allClustersOnDB, err := handleLocalPolicyStatus(tx,
			leafHubName, clusterPerPolicyBundle.PolicyID,
			clusterPerPolicyBundle.CompliantClusters, allClustersOnDB, database.Compliant,
			clusterSetFromDB.GetClusters(database.Compliant))
		if err != nil {
			return err
		}

		// handle non compliant clusters of the policy
		allClustersOnDB, err = handleLocalPolicyStatus(tx,
			leafHubName, clusterPerPolicyBundle.PolicyID,
			clusterPerPolicyBundle.NonCompliantClusters, allClustersOnDB, database.NonCompliant,
			clusterSetFromDB.GetClusters(database.NonCompliant))
		if err != nil {
			return err
		}

		// handle unknown compliance clusters of the policy
		allClustersOnDB, err = handleLocalPolicyStatus(tx,
			leafHubName, clusterPerPolicyBundle.PolicyID,
			clusterPerPolicyBundle.UnknownComplianceClusters, allClustersOnDB, database.Unknown,
			clusterSetFromDB.GetClusters(database.Unknown))
		if err != nil {
			return err
		}

		// delete compliance status rows in the db that were not sent in the bundle
		for _, name := range allClustersOnDB.ToSlice() {
			clusterName, ok := name.(string)
			if !ok {
				continue
			}
			err := tx.Where(&models.LocalStatusCompliance{
				LeafHubName: leafHubName,
				PolicyID:    clusterPerPolicyBundle.PolicyID,
				ClusterName: clusterName,
			}).Delete(&models.LocalStatusCompliance{}).Error
			if err != nil {
				return err
			}
		}
		// return nil will commit the whole transaction
		return nil
	})
	return err
}

func handleLocalPolicyStatus(tx *gorm.DB, leafHub, policyID string, bundleClusters []string,
	allClusterFromDB set.Set, complianceStatus database.ComplianceStatus, typedClusters set.Set,
) (set.Set, error) {
	for _, clusterName := range bundleClusters {
		// if the policy on cluster not exist or compliance is updated
		if !allClusterFromDB.Contains(clusterName) || !typedClusters.Contains(clusterName) {
			err := tx.Save(&models.LocalStatusCompliance{
				LeafHubName: leafHub,
				PolicyID:    policyID,
				ClusterName: clusterName,
				Error:       database.ErrorNone,
				Compliance:  complianceStatus,
			}).Error
			if err != nil {
				return nil, err
			}
		}

		// either way if status was updated or not, remove from allClustersFromDB to mark this cluster as handled
		allClusterFromDB.Remove(clusterName)
	}
	return allClusterFromDB, nil
}

func (syncer *PoliciesDBSyncer) handleCompleteLocalStatusComplianceBundle(ctx context.Context,
	bundle status.Bundle,
) error {
	logBundleHandlingMessage(syncer.log, bundle, startBundleHandlingMessage)
	leafHubName := bundle.GetLeafHubName()
	db := database.GetGorm()

	// policyID: { compliance: (cluster1, cluster2), nonCompliance: (cluster3, cluster4), unknowns: (cluster5) }
	allPolicyComplianceRowsFromDB, err := getAllLocalPolicyClusterSets(db,
		"leaf_hub_name = ? AND compliance <> ?",
		leafHubName, database.Compliant)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, object := range bundle.GetObjects() { // every object in bundle is policy compliance status
			policyComplianceStatus, ok := object.(*status.PolicyCompleteComplianceStatus)
			if !ok {
				continue // do not handle objects other than PolicyComplianceStatus
			}
			// nonCompliantClusters includes both non Compliant and Unknown clusters
			nonComplianceClusterSetsFromDB, policyExistsInDB :=
				allPolicyComplianceRowsFromDB[policyComplianceStatus.PolicyID]
			if !policyExistsInDB {
				nonComplianceClusterSetsFromDB = NewPolicyClusterSets()
			}
			allNonComplianceClusters := nonComplianceClusterSetsFromDB.GetAllClusters()

			// update in db batch the non Compliant clusters as it was reported by leaf hub
			for _, clusterName := range policyComplianceStatus.NonCompliantClusters { // go over bundle non compliant clusters
				if !nonComplianceClusterSetsFromDB.GetClusters(
					database.NonCompliant).Contains(clusterName) {
					err := updateLocalStatusCompliance(tx,
						policyComplianceStatus.PolicyID, leafHubName, clusterName,
						database.NonCompliant)
					if err != nil {
						return err
					}
				} // if different need to update, otherwise no need to do anything.
				allNonComplianceClusters.Remove(clusterName) // mark cluster as handled
			}

			// update in db batch the unknown clusters as it was reported by leaf hub
			for _, clusterName := range policyComplianceStatus.UnknownComplianceClusters { // go over bundle unknown clusters
				if !nonComplianceClusterSetsFromDB.GetClusters(database.Unknown).Contains(clusterName) {
					err := updateLocalStatusCompliance(tx,
						policyComplianceStatus.PolicyID, leafHubName,
						clusterName, database.Unknown)
					if err != nil {
						return err
					}
				} // if different need to update, otherwise no need to do anything.
				allNonComplianceClusters.Remove(clusterName) // mark cluster as handled
			}

			for _, name := range allNonComplianceClusters.ToSlice() {
				clusterName, ok := name.(string)
				if !ok {
					continue
				}
				err := updateLocalStatusCompliance(tx,
					policyComplianceStatus.PolicyID, leafHubName, clusterName,
					database.Compliant)
				if err != nil {
					return err
				}
			}

			// for policies that are found in the db but not in the bundle - all clusters are Compliant (implicitly)
			delete(allPolicyComplianceRowsFromDB, policyComplianceStatus.PolicyID)
		}

		// update policies not in the bundle - all is Compliant
		for policyID := range allPolicyComplianceRowsFromDB {
			ret := tx.Model(&models.LocalStatusCompliance{}).Where("policy_id = ? AND leaf_hub_name = ?",
				policyID, leafHubName).Updates(&models.LocalStatusCompliance{Compliance: database.Compliant})
			if ret.Error != nil {
				return ret.Error
			}
		}
		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to handle complete compliance bundle - %w", err)
	}

	logBundleHandlingMessage(syncer.log, bundle, finishBundleHandlingMessage)
	return nil
}

func getAllLocalPolicyClusterSets(db *gorm.DB, query interface{}, args ...interface{}) (
	map[string]*PolicyClustersSets, error,
) {
	var compliancesFromDB []models.LocalStatusCompliance
	err := db.Where(query, args...).
		Find(&compliancesFromDB).Error
	if err != nil {
		return nil, err
	}

	// policyID: { compliance: (cluster1, cluster2), nonCompliance: (cluster3, cluster4), unknowns: (cluster5) }
	allPolicyComplianceRowsFromDB := make(map[string]*PolicyClustersSets)
	for _, compliance := range compliancesFromDB {
		if _, ok := allPolicyComplianceRowsFromDB[compliance.PolicyID]; !ok {
			allPolicyComplianceRowsFromDB[compliance.PolicyID] = NewPolicyClusterSets()
		}
		allPolicyComplianceRowsFromDB[compliance.PolicyID].AddCluster(
			compliance.ClusterName, compliance.Compliance)
	}
	return allPolicyComplianceRowsFromDB, nil
}

func updateLocalStatusCompliance(tx *gorm.DB, policyID string, leafHubName string, clusterName string,
	compliance database.ComplianceStatus,
) error {
	return tx.Model(&models.LocalStatusCompliance{}).Where(&models.LocalStatusCompliance{
		PolicyID:    policyID,
		ClusterName: clusterName,
		LeafHubName: leafHubName,
	}).Updates(&models.LocalStatusCompliance{
		Compliance: compliance,
	}).Error
}
