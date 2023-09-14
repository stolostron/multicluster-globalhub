package task

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"

	"github.com/stolostron/multicluster-global-hub/pkg/database"
	"github.com/stolostron/multicluster-global-hub/pkg/database/models"
)

var _ = Describe("data retention job", Ordered, func() {
	expiredPartitionTables := map[string]bool{}
	currentTime := time.Now()
	duration := time.Duration(18) * 30 * 24 * time.Hour

	minTime := currentTime.Add(-duration)
	expirationTime := currentTime.Add(-duration).AddDate(0, -1, 0)
	maxTime := currentTime.AddDate(0, 1, 0)

	BeforeAll(func() {
		By("Creating expired partition table in the database")
		for _, tableName := range partitionTables {
			err := createPartitionTable(tableName, expirationTime)
			Expect(err).ToNot(HaveOccurred())
			expiredPartitionTables[fmt.Sprintf("%s_%s", tableName,
				expirationTime.Format(partitionDateFormat))] = false
		}

		By("Create the min partition table in the database")
		for _, tableName := range partitionTables {
			err := createPartitionTable(tableName, minTime)
			Expect(err).ToNot(HaveOccurred())
		}

		By("Check whether the expired tables are created")
		Eventually(func() error {
			var tables []models.Table
			result := db.Raw("SELECT schemaname as schema_name, tablename as table_name FROM pg_tables").Find(&tables)
			if result.Error != nil {
				return result.Error
			}
			for _, table := range tables {
				gotTable := fmt.Sprintf("%s.%s", table.Schema, table.Table)
				if _, ok := expiredPartitionTables[gotTable]; ok {
					fmt.Println("the expired partition table is created: ", gotTable)
					expiredPartitionTables[gotTable] = true
				}
			}
			for key, val := range expiredPartitionTables {
				if !val {
					return fmt.Errorf("table %s is not created", key)
				}
			}
			return nil
		}, 10*time.Second, 1*time.Second).ShouldNot(HaveOccurred())

		By("Create soft deleted recorded in the databse")
		for _, tableName := range retentionTables {
			err := createRetentionData(tableName, expirationTime)
			Expect(err).ToNot(HaveOccurred())
		}

		for _, table := range retentionTables {
			By(fmt.Sprintf("Check whether the record was created in table %s", table))
			Eventually(func() error {
				rows, err := db.Raw(fmt.Sprintf(`SELECT leaf_hub_name, deleted_at FROM %s WHERE DELETED_AT <= '%s'`,
					table, expirationTime.Format(timeFormat))).Rows()
				if err != nil {
					return fmt.Errorf("error reading from table %s due to: %v", table, err)
				}
				defer rows.Close()

				if !rows.Next() {
					return fmt.Errorf("The record was not exists in table %s due to: %v", table, err)
				}

				fmt.Println("the deleted record is created: ", table)
				return nil
			}, 10*time.Second, 1*time.Second).Should(BeNil())
		}
	})

	It("the data retention job should work", func() {
		By("Create the data retention job")
		s := gocron.NewScheduler(time.UTC)
		_, err := s.Every(1).Week().DoWithJobDetails(DataRetention, ctx, pool, duration)
		Expect(err).ToNot(HaveOccurred())
		s.StartAsync()
		defer s.Clear()

		By("Check whether the expired tables are deleted")
		Eventually(func() error {
			var tables []models.Table
			result := db.Raw("SELECT schemaname as schema_name, tablename as table_name FROM pg_tables").Find(&tables)
			if result.Error != nil {
				return result.Error
			}

			count := 0
			for _, table := range tables {
				gotTable := fmt.Sprintf("%s.%s", table.Schema, table.Table)
				if _, ok := expiredPartitionTables[gotTable]; ok {
					fmt.Println("deleting the expired partition table: ", gotTable)
					count++
				}
			}
			if count > 0 {
				return fmt.Errorf("the expired tables hasn't been deleted")
			}
			return nil
		}, 10*time.Second, 1*time.Second).ShouldNot(HaveOccurred())

		for _, table := range retentionTables {
			By(fmt.Sprintf("Check whether the record were deleted in table %s", table))
			Eventually(func() error {
				rows, err := db.Raw(fmt.Sprintf(`SELECT leaf_hub_name, deleted_at FROM %s WHERE DELETED_AT <= '%s'`,
					table, expirationTime.Format(timeFormat))).Rows()
				if err != nil {
					return fmt.Errorf("error reading from table %s due to: %v", table, err)
				}
				defer rows.Close()
				if rows.Next() {
					return fmt.Errorf("The record was not exists in table %s due to: %v", table, err)
				}

				fmt.Println("deleting the expired record in table: ", table)
				return nil
			}, 10*time.Second, 1*time.Second).Should(BeNil())
		}
	})

	It("the data retention should log the job execution", func() {
		db := database.GetGorm()
		logs := []models.DataRetentionJobLog{}

		Eventually(func() error {
			result := db.Find(&logs)
			if result.Error != nil {
				return result.Error
			}
			if len(logs) < 6 {
				return fmt.Errorf("the logs are not enough")
			}
			return nil
		}, 10*time.Second, 1*time.Second).ShouldNot(HaveOccurred())
		for _, log := range logs {
			fmt.Printf("table_name(%s) | min(%s) | max(%s) | min_deletion(%s) \n", log.Name, log.MinPartition,
				log.MaxPartition, log.MinDeletion.Format(dateFormat))
			for _, tableName := range partitionTables {
				if log.Name == tableName {
					Expect(log.MinPartition).To(ContainSubstring(minTime.Format(partitionDateFormat)))
					Expect(log.MaxPartition).To(ContainSubstring(maxTime.Format(partitionDateFormat)))
				}
			}
		}
	})
})

func createPartitionTable(tableName string, date time.Time) error {
	db := database.GetGorm()
	result := db.Exec(`SELECT create_monthly_range_partitioned_table(?, ?)`, tableName, date.Format(dateFormat))
	if result.Error != nil {
		return fmt.Errorf("failed to create partition table %s: %w", tableName, result.Error)
	}
	return nil
}

func createRetentionData(tableName string, date time.Time) error {
	var result *gorm.DB

	db := database.GetGorm()

	switch tableName {
	case "status.managed_clusters":
		uid := uuid.New().String()
		mcPayload := fmt.Sprintf(`
		{
			"kind": "ManagedCluster", 
			"spec": {
				"hubAcceptsClient": true, 
				"leaseDurationSeconds": 60
				}, 
			"metadata": {
				"uid": %s, 
				"name": "leafhub1"
			}, 
			"apiVersion": "cluster.open-cluster-management.io/v1"
		}`, uid)
		result = db.Exec(
			fmt.Sprintf(`INSERT INTO status.managed_clusters (leaf_hub_name, cluster_id, payload, error, 
				created_at, updated_at, deleted_at) VALUES ("leafhub1", "%s", "%s", "none", "%s", "%s", "%s")`, uid,
				mcPayload, date.Format(timeFormat), date.Format(timeFormat), date.Format(timeFormat)))

	case "status.leaf_hubs":
		result = db.Exec(
			fmt.Sprintf(`INSERT INTO status.leaf_hubs (leaf_hub_name, payload, created_at, updated_at, deleted_at) 
			VALUES ('leafhub1', '{"consoleURL": "https://leafhub1.com", "leafHubName": "leafhub1"}', '%s', '%s', '%s')`,
				date.Format(timeFormat), date.Format(timeFormat), date.Format(timeFormat)))

	case "local_spec.policies":
		uid := uuid.New().String()
		policyPayload := fmt.Sprintf(`
		{
			"kind": "Policy", 
			"spec": {}, 
			"metadata": {
				"uid": %s, 
				"name": "policy1", 
				"namespace": "default"
			}, 
			"apiVersion": "policy.open-cluster-management.io/v1"
		}`, uid)
		result = db.Exec(
			fmt.Sprintf(`INSERT INTO local_spec.policies (policy_id, leaf_hub_name, payload, created_at, 
				updated_at, deleted_at) VALUES ("%s", "leafhub1", "%s", "%s", "%s", "%s")`, uid, policyPayload,
				date.Format(timeFormat), date.Format(timeFormat), date.Format(timeFormat)))
	}
	if result.Error != nil {
		return fmt.Errorf("failed to create retention data in table %s due to: %w", tableName, result.Error)
	}
	return nil
}
