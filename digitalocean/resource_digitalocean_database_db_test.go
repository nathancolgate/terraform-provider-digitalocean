package digitalocean

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDigitalOceanDatabaseDB_Basic(t *testing.T) {
	var databaseDB godo.DatabaseDB
	databaseClusterName := fmt.Sprintf("foobar-test-terraform-%s", acctest.RandString(10))
	databaseDBName := fmt.Sprintf("foobar-test-db-terraform-%s", acctest.RandString(10))
	databaseDBNameUpdated := databaseDBName + "-up"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDigitalOceanDatabaseDBDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckDigitalOceanDatabaseDBConfigBasic, databaseClusterName, databaseDBName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDigitalOceanDatabaseDBExists("digitalocean_database_db.foobar_db", &databaseDB),
					testAccCheckDigitalOceanDatabaseDBAttributes(&databaseDB, databaseDBName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_db.foobar_db", "name", databaseDBName),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckDigitalOceanDatabaseDBConfigBasic, databaseClusterName, databaseDBNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDigitalOceanDatabaseDBExists("digitalocean_database_db.foobar_db", &databaseDB),
					testAccCheckDigitalOceanDatabaseDBNotExists("digitalocean_database_db.foobar_db", databaseDBName),
					testAccCheckDigitalOceanDatabaseDBAttributes(&databaseDB, databaseDBNameUpdated),
					resource.TestCheckResourceAttr(
						"digitalocean_database_db.foobar_db", "name", databaseDBNameUpdated),
				),
			},
		},
	})
}

func testAccCheckDigitalOceanDatabaseDBDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CombinedConfig).godoClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "digitalocean_database_db" {
			continue
		}
		clusterID := rs.Primary.Attributes["cluster_id"]
		name := rs.Primary.Attributes["name"]

		// Try to find the database DB
		_, _, err := client.Databases.GetDB(context.Background(), clusterID, name)

		if err == nil {
			return fmt.Errorf("Database DB still exists")
		}
	}

	return nil
}

func testAccCheckDigitalOceanDatabaseDBExists(n string, databaseDB *godo.DatabaseDB) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Database DB ID is set")
		}

		client := testAccProvider.Meta().(*CombinedConfig).godoClient()
		clusterID := rs.Primary.Attributes["cluster_id"]
		name := rs.Primary.Attributes["name"]

		foundDatabaseDB, _, err := client.Databases.GetDB(context.Background(), clusterID, name)

		if err != nil {
			return err
		}

		if foundDatabaseDB.Name != name {
			return fmt.Errorf("Database DB not found")
		}

		*databaseDB = *foundDatabaseDB

		return nil
	}
}

func testAccCheckDigitalOceanDatabaseDBNotExists(n string, databaseDBName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Database DB ID is set")
		}

		client := testAccProvider.Meta().(*CombinedConfig).godoClient()
		clusterID := rs.Primary.Attributes["cluster_id"]

		_, resp, err := client.Databases.GetDB(context.Background(), clusterID, databaseDBName)

		if err != nil && resp.StatusCode != http.StatusNotFound {
			return err
		}

		if err == nil {
			return fmt.Errorf("Database DB %s still exists", databaseDBName)
		}

		return nil
	}
}

func testAccCheckDigitalOceanDatabaseDBAttributes(databaseDB *godo.DatabaseDB, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if databaseDB.Name != name {
			return fmt.Errorf("Bad name: %s", databaseDB.Name)
		}

		return nil
	}
}

const testAccCheckDigitalOceanDatabaseDBConfigBasic = `
resource "digitalocean_database_cluster" "foobar" {
	name       = "%s"
	engine     = "pg"
	version    = "11"
	size       = "db-s-1vcpu-1gb"
	region     = "nyc1"
	node_count = 1

	maintenance_window {
        day  = "friday"
        hour = "13:00:00"
	}
}

resource "digitalocean_database_db" "foobar_db" {
  cluster_id = "${digitalocean_database_cluster.foobar.id}"
  name       = "%s"
}`
