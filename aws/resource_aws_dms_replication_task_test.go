package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDmsReplicationTaskBasic(t *testing.T) {
	resourceName := "aws_dms_replication_task.dms_replication_task"
	randId := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsReplicationTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsReplicationTaskConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsReplicationTaskExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: dmsReplicationTaskConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsReplicationTaskExists(resourceName),
				),
			},
		},
	})
}

func TestAccAwsDmsReplicationTaskHandleLifecycle(t *testing.T) {
	resourceName := "aws_dms_replication_task.dms_replication_task"
	randId := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsReplicationTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsReplicationTaskConfigLifecycle(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsReplicationTaskExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
				),
			},
		},
	})
}

func checkDmsReplicationTaskExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn
		resp, err := conn.DescribeReplicationTasks(&dms.DescribeReplicationTasksInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("replication-task-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			return err
		}

		if resp.ReplicationTasks == nil {
			return fmt.Errorf("DMS replication task error: %v", err)
		}
		return nil
	}
}

func dmsReplicationTaskDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dms_replication_task" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn
		resp, err := conn.DescribeReplicationTasks(&dms.DescribeReplicationTasksInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("replication-task-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			return nil
		}

		if resp != nil && len(resp.ReplicationTasks) > 0 {
			return fmt.Errorf("DMS replication task still exists: %v", err)
		}
	}

	return nil
}

func dmsReplicationTaskConfig(randId string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "dms_vpc" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-dms-replication-task"
	}
}

resource "aws_subnet" "dms_subnet_1" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.dms_vpc.id}"
	tags = {
		Name = "tf-acc-dms-replication-task-1"
	}
	depends_on = ["aws_vpc.dms_vpc"]
}

resource "aws_subnet" "dms_subnet_2" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.dms_vpc.id}"
	tags = {
		Name = "tf-acc-dms-replication-task-2"
	}
	depends_on = ["aws_vpc.dms_vpc"]
}

resource "aws_dms_endpoint" "dms_endpoint_source" {
	database_name = "tf-test-dms-db"
	endpoint_id = "tf-test-dms-endpoint-source-%[1]s"
	endpoint_type = "source"
	engine_name = "aurora"
	server_name = "tf-test-cluster.cluster-xxxxxxx.us-west-2.rds.amazonaws.com"
	port = 3306
	username = "tftest"
	password = "tftest"
}

resource "aws_dms_endpoint" "dms_endpoint_target" {
	database_name = "tf-test-dms-db"
	endpoint_id = "tf-test-dms-endpoint-target-%[1]s"
	endpoint_type = "target"
	engine_name = "aurora"
	server_name = "tf-test-cluster.cluster-xxxxxxx.us-west-2.rds.amazonaws.com"
	port = 3306
	username = "tftest"
	password = "tftest"
}

resource "aws_dms_replication_subnet_group" "dms_replication_subnet_group" {
	replication_subnet_group_id = "tf-test-dms-replication-subnet-group-%[1]s"
	replication_subnet_group_description = "terraform test for replication subnet group"
	subnet_ids = ["${aws_subnet.dms_subnet_1.id}", "${aws_subnet.dms_subnet_2.id}"]
}

resource "aws_dms_replication_instance" "dms_replication_instance" {
	allocated_storage = 5
	auto_minor_version_upgrade = true
	replication_instance_class = "dms.t2.micro"
	replication_instance_id = "tf-test-dms-replication-instance-%[1]s"
	preferred_maintenance_window = "sun:00:30-sun:02:30"
	publicly_accessible = false
	replication_subnet_group_id = "${aws_dms_replication_subnet_group.dms_replication_subnet_group.replication_subnet_group_id}"
}

resource "aws_dms_replication_task" "dms_replication_task" {
	migration_type = "full-load-and-cdc"
	replication_instance_arn = "${aws_dms_replication_instance.dms_replication_instance.replication_instance_arn}"
	replication_task_id = "tf-test-dms-replication-task-%[1]s"
	replication_task_settings = "{\"TargetMetadata\":{\"TargetSchema\":\"\",\"SupportLobs\":true,\"FullLobMode\":false,\"LobChunkSize\":0,\"LimitedSizeLobMode\":true,\"LobMaxSize\":32,\"LoadMaxFileSize\":0,\"ParallelLoadThreads\":0,\"BatchApplyEnabled\":false},\"FullLoadSettings\":{\"FullLoadEnabled\":true,\"ApplyChangesEnabled\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"CreatePkAfterFullLoad\":false,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"ResumeEnabled\":false,\"ResumeMinTableSize\":100000,\"ResumeOnlyClusteredPKTables\":true,\"MaxFullLoadSubTasks\":8,\"TransactionConsistencyTimeout\":600,\"CommitRate\":10000},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}],\"CloudWatchLogGroup\":null,\"CloudWatchLogStream\":null},\"ControlTablesSettings\":{\"historyTimeslotInMinutes\":5,\"ControlSchema\":\"\",\"HistoryTimeslotInMinutes\":5,\"HistoryTableEnabled\":false,\"SuspendedTablesTableEnabled\":false,\"StatusTableEnabled\":false},\"StreamBufferSettings\":{\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8,\"CtrlStreamBufferSizeInMB\":5},\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true,\"HandleSourceTableAltered\":true},\"ErrorBehavior\":{\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorEscalationCount\":0,\"TableErrorPolicy\":\"SUSPEND_TABLE\",\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorEscalationCount\":0,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorEscalationCount\":0,\"FullLoadIgnoreConflicts\":true},\"ChangeProcessingTuning\":{\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMin\":1,\"BatchApplyTimeoutMax\":30,\"BatchApplyMemoryLimit\":500,\"BatchSplitSize\":0,\"MinTransactionSize\":1000,\"CommitTimeout\":1,\"MemoryLimitTotal\":1024,\"MemoryKeepTime\":60,\"StatementCacheSize\":50}}"
	source_endpoint_arn = "${aws_dms_endpoint.dms_endpoint_source.endpoint_arn}"
	table_mappings = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
	tags = {
		Name = "tf-test-dms-replication-task-%[1]s"
		Update = "to-update"
		Remove = "to-remove"
	}
	target_endpoint_arn = "${aws_dms_endpoint.dms_endpoint_target.endpoint_arn}"
}
`, randId)
}

func dmsReplicationTaskConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "dms_vpc" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-dms-replication-task"
	}
}

resource "aws_subnet" "dms_subnet_1" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.dms_vpc.id}"
	tags = {
		Name = "tf-acc-dms-replication-task-1"
	}
	depends_on = ["aws_vpc.dms_vpc"]
}

resource "aws_subnet" "dms_subnet_2" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.dms_vpc.id}"
	tags = {
		Name = "tf-acc-dms-replication-task-2"
	}
	depends_on = ["aws_vpc.dms_vpc"]
}

resource "aws_dms_endpoint" "dms_endpoint_source" {
	database_name = "tf-test-dms-db"
	endpoint_id = "tf-test-dms-endpoint-source-%[1]s"
	endpoint_type = "source"
	engine_name = "aurora"
	server_name = "tf-test-cluster.cluster-xxxxxxx.us-west-2.rds.amazonaws.com"
	port = 3306
	username = "tftest"
	password = "tftest"
}

resource "aws_dms_endpoint" "dms_endpoint_target" {
	database_name = "tf-test-dms-db"
	endpoint_id = "tf-test-dms-endpoint-target-%[1]s"
	endpoint_type = "target"
	engine_name = "aurora"
	server_name = "tf-test-cluster.cluster-xxxxxxx.us-west-2.rds.amazonaws.com"
	port = 3306
	username = "tftest"
	password = "tftest"
}

resource "aws_dms_replication_subnet_group" "dms_replication_subnet_group" {
	replication_subnet_group_id = "tf-test-dms-replication-subnet-group-%[1]s"
	replication_subnet_group_description = "terraform test for replication subnet group"
	subnet_ids = ["${aws_subnet.dms_subnet_1.id}", "${aws_subnet.dms_subnet_2.id}"]
}

resource "aws_dms_replication_instance" "dms_replication_instance" {
	allocated_storage = 5
	auto_minor_version_upgrade = true
	replication_instance_class = "dms.t2.micro"
	replication_instance_id = "tf-test-dms-replication-instance-%[1]s"
	preferred_maintenance_window = "sun:00:30-sun:02:30"
	publicly_accessible = false
	replication_subnet_group_id = "${aws_dms_replication_subnet_group.dms_replication_subnet_group.replication_subnet_group_id}"
}

resource "aws_dms_replication_task" "dms_replication_task" {
	migration_type = "full-load"
	replication_instance_arn = "${aws_dms_replication_instance.dms_replication_instance.replication_instance_arn}"
	replication_task_id = "tf-test-dms-replication-task-%[1]s"
	replication_task_settings = "{\"TargetMetadata\":{\"TargetSchema\":\"\",\"SupportLobs\":true,\"FullLobMode\":false,\"LobChunkSize\":0,\"LimitedSizeLobMode\":true,\"LobMaxSize\":32,\"LoadMaxFileSize\":0,\"ParallelLoadThreads\":0,\"BatchApplyEnabled\":false},\"FullLoadSettings\":{\"FullLoadEnabled\":true,\"ApplyChangesEnabled\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"CreatePkAfterFullLoad\":false,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"ResumeEnabled\":false,\"ResumeMinTableSize\":100000,\"ResumeOnlyClusteredPKTables\":true,\"MaxFullLoadSubTasks\":7,\"TransactionConsistencyTimeout\":600,\"CommitRate\":10000},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}],\"CloudWatchLogGroup\":null,\"CloudWatchLogStream\":null},\"ControlTablesSettings\":{\"historyTimeslotInMinutes\":5,\"ControlSchema\":\"\",\"HistoryTimeslotInMinutes\":5,\"HistoryTableEnabled\":false,\"SuspendedTablesTableEnabled\":false,\"StatusTableEnabled\":false},\"StreamBufferSettings\":{\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8,\"CtrlStreamBufferSizeInMB\":5},\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true,\"HandleSourceTableAltered\":true},\"ErrorBehavior\":{\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorEscalationCount\":0,\"TableErrorPolicy\":\"SUSPEND_TABLE\",\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorEscalationCount\":0,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorEscalationCount\":0,\"FullLoadIgnoreConflicts\":true},\"ChangeProcessingTuning\":{\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMin\":1,\"BatchApplyTimeoutMax\":30,\"BatchApplyMemoryLimit\":500,\"BatchSplitSize\":0,\"MinTransactionSize\":1000,\"CommitTimeout\":1,\"MemoryLimitTotal\":1024,\"MemoryKeepTime\":60,\"StatementCacheSize\":50}}"
	source_endpoint_arn = "${aws_dms_endpoint.dms_endpoint_source.endpoint_arn}"
	table_mappings = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
	tags = {
		Name = "tf-test-dms-replication-task-%[1]s"
		Update = "updated"
		Add = "added"
	}
	target_endpoint_arn = "${aws_dms_endpoint.dms_endpoint_target.endpoint_arn}"
}
`, randId)
}

func dmsReplicationTaskConfigLifecycle(randId string) string {
	return fmt.Sprintf(`
variable "db_user" {
  default = "root"
}

variable "db_pass" {
  default = "abcd1234"
}

resource "aws_vpc" "dms_vpc" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags {
    Name = "tf-test-dms-vpc-%[1]s"
  }
}

resource "aws_internet_gateway" "public" {
  vpc_id = "${aws_vpc.dms_vpc.id}"
}

resource "aws_route" "public" {
  route_table_id         = "${aws_vpc.dms_vpc.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.public.id}"
}

resource "aws_subnet" "dms_subnet_1" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  vpc_id            = "${aws_vpc.dms_vpc.id}"

  tags {
    Name = "tf-test-dms-subnet-%[1]s"
  }

  depends_on = ["aws_vpc.dms_vpc"]
}

resource "aws_subnet" "dms_subnet_2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2b"
  vpc_id            = "${aws_vpc.dms_vpc.id}"

  tags {
    Name = "tf-test-dms-subnet-%[1]s"
  }

  depends_on = ["aws_vpc.dms_vpc"]
}

resource "aws_dms_endpoint" "dms_endpoint_source" {
  database_name = "test_dms"
  endpoint_id   = "tf-test-dms-endpoint-source-%[1]s"
  endpoint_type = "source"
  engine_name   = "mysql"
  server_name   = "${aws_db_instance.source_mysql.address}"
  port          = 3306
  username      = "${var.db_user}"
  password      = "${var.db_pass}"
}

resource "aws_dms_endpoint" "dms_endpoint_target" {
  database_name = "test_dms"
  endpoint_id   = "tf-test-dms-endpoint-target-%[1]s"
  endpoint_type = "target"
  engine_name   = "mysql"
  server_name   = "${aws_db_instance.target_mysql.address}"
  port          = 3306
  username      = "${var.db_user}"
  password      = "${var.db_pass}"
}

resource "aws_dms_replication_subnet_group" "dms_replication_subnet_group" {
  replication_subnet_group_id          = "tf-test-dms-replication-subnet-group-%[1]s"
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = ["${aws_subnet.dms_subnet_1.id}", "${aws_subnet.dms_subnet_2.id}"]
}

resource "aws_dms_replication_instance" "dms_replication_instance" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.t2.micro"
  replication_instance_id      = "tf-test-dms-replication-instance-%[1]s"
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = "${aws_dms_replication_subnet_group.dms_replication_subnet_group.replication_subnet_group_id}"

  depends_on = ["aws_instance.bastion"]
}

resource "aws_dms_replication_task" "dms_replication_task" {
  migration_type = "full-load"

  handle_task_lifecycle     = true
  replication_instance_arn  = "${aws_dms_replication_instance.dms_replication_instance.replication_instance_arn}"
  replication_task_id       = "tf-test-dms-replication-task-%[1]s"
  replication_task_settings = "{\"TargetMetadata\":{\"TargetSchema\":\"\",\"SupportLobs\":true,\"FullLobMode\":false,\"LobChunkSize\":0,\"LimitedSizeLobMode\":true,\"LobMaxSize\":32,\"LoadMaxFileSize\":0,\"ParallelLoadThreads\":0,\"BatchApplyEnabled\":false},\"FullLoadSettings\":{\"FullLoadEnabled\":true,\"ApplyChangesEnabled\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"CreatePkAfterFullLoad\":false,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"ResumeEnabled\":false,\"ResumeMinTableSize\":100000,\"ResumeOnlyClusteredPKTables\":true,\"MaxFullLoadSubTasks\":7,\"TransactionConsistencyTimeout\":600,\"CommitRate\":10000},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}],\"CloudWatchLogGroup\":null,\"CloudWatchLogStream\":null},\"ControlTablesSettings\":{\"historyTimeslotInMinutes\":5,\"ControlSchema\":\"\",\"HistoryTimeslotInMinutes\":5,\"HistoryTableEnabled\":false,\"SuspendedTablesTableEnabled\":false,\"StatusTableEnabled\":false},\"StreamBufferSettings\":{\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8,\"CtrlStreamBufferSizeInMB\":5},\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true,\"HandleSourceTableAltered\":true},\"ErrorBehavior\":{\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorEscalationCount\":0,\"TableErrorPolicy\":\"SUSPEND_TABLE\",\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorEscalationCount\":0,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorEscalationCount\":0,\"FullLoadIgnoreConflicts\":true},\"ChangeProcessingTuning\":{\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMin\":1,\"BatchApplyTimeoutMax\":30,\"BatchApplyMemoryLimit\":500,\"BatchSplitSize\":0,\"MinTransactionSize\":1000,\"CommitTimeout\":1,\"MemoryLimitTotal\":1024,\"MemoryKeepTime\":60,\"StatementCacheSize\":50}}"
  source_endpoint_arn       = "${aws_dms_endpoint.dms_endpoint_source.endpoint_arn}"
  table_mappings            = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"test_dms\",\"table-name\":\"test_table1\"},\"rule-action\":\"include\"}]}"

  tags {
    Name   = "tf-test-dms-replication-task"
    Update = "updated"
    Add    = "added"
  }

  target_endpoint_arn = "${aws_dms_endpoint.dms_endpoint_target.endpoint_arn}"
}

# setup target & source dbs
resource "aws_db_instance" "source_mysql" {
  allocated_storage       = 10
  engine                  = "mysql"
  engine_version          = "5.6.37"
  identifier              = "dms-test-source-%[1]s"
  instance_class          = "db.t2.micro"
  storage_type            = "gp2"
  skip_final_snapshot     = true
  username                = "${var.db_user}"
  password                = "${var.db_pass}"
  port                    = "3306"
  backup_retention_period = 1
  multi_az                = false
  publicly_accessible     = true
  vpc_security_group_ids  = ["${aws_security_group.allow_dms.id}"]
  db_subnet_group_name    = "${aws_db_subnet_group.rds.id}"
  parameter_group_name    = "default.mysql5.6"
}

resource "aws_db_instance" "target_mysql" {
  allocated_storage       = 10
  engine                  = "mysql"
  engine_version          = "5.6.37"
  identifier              = "dms-test-target-%[1]s"
  instance_class          = "db.t2.micro"
  storage_type            = "gp2"
  name                    = "test_dms"
  skip_final_snapshot     = true
  username                = "${var.db_user}"
  password                = "${var.db_pass}"
  port                    = "3306"
  backup_retention_period = 1
  multi_az                = false
  publicly_accessible     = true
  vpc_security_group_ids  = ["${aws_security_group.allow_dms.id}"]
  db_subnet_group_name    = "${aws_db_subnet_group.rds.id}"
  parameter_group_name    = "default.mysql5.6"
}

resource "aws_db_subnet_group" "rds" {
  name        = "db-subnet-group-test-%[1]s"
  description = "RDS subnet group for eu-west region"
  subnet_ids  = ["${aws_subnet.dms_subnet_1.id}", "${aws_subnet.dms_subnet_2.id}"]
}

resource "aws_security_group" "allow_dms" {
  name        = "allow_all_rds-%[1]s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.dms_vpc.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# setup bastion for loading dms
resource "aws_instance" "bastion" {
  ami                         = "ami-32d8124a"
  instance_type               = "t2.micro"
  vpc_security_group_ids      = ["${aws_security_group.allow_dms.id}"]
  subnet_id                   = "${aws_subnet.dms_subnet_1.id}"
  associate_public_ip_address = true
  key_name                    = "${aws_key_pair.deployer.key_name}"

  provisioner "remote-exec" {
    inline = [
      "sudo yum install mysql -y",
      "mysql -u ${var.db_user} -h ${aws_db_instance.source_mysql.address} -p${var.db_pass} -e 'CREATE DATABASE test_dms; USE test_dms; CREATE TABLE test_table1 (pri_key varchar(2) NOT NULL, some_field varchar(10) DEFAULT NULL, PRIMARY KEY (pri_key)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4; INSERT INTO test_table1 VALUES (\"AA\",\"Some Value\");'",
    ]

    connection {
      type        = "ssh"
      user        = "ec2-user"
      private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAspEROVtau2HOTgR+5TEYdt7ohiiviKMURlFzgE6t2ERpRhjU
xG6Dnpdn0oXy/9BFU9wWl5xq+VNMWXn4GCHAIOlbb5Zo7E/W7jwvPAOfdeY1KLCj
jlm0q2RZzRhuHMIQLFP6WVcjOPQlyU+Q4l7XQFlUcpYuzkC0iCJhSRfJQ6Le7c3k
Xgnm1WjwGiGilQNlnQsra0zTXc7jTImv4qxaRGUKtp4bQ2m/NHAZFlctjm56ZvOc
WUuwGv7t1HOl58hdgf2kSiyk1h6siBYdo3zcGyouo+RTyEeZrxxTLG4IcQKQBa7l
gv8T19dPUl02AJzaTeiLHQU3hOwbA+ls+2MsYQIDAQABAoIBAGwcBC+TraURHBSE
CEe+p68gWesPquawxU+ldKZT/FCZapsz4W1j83AK/qKo0mwqri6Na2gzHVkCI5Fw
lNIXbPkAD4nJqJCZ7eiiq35MOzjoPXr7JqrCiO3TfcL8bX4fyCbuWP3KEdsjhdUR
xQgon22oJ8aQQppA9owNNJVKP2Igr8QwyZbRB2Yi4QSvDpS1LMvoS6vK3lxfh9p3
z5qFNNdlZ+Y3i0jl3Z0mDrWRg1M25C9MHOQfYromc3fWJ3O1XEo3BjGgNxmjSH47
fVlmu/wmTeNg5DCVvll5Lag7lDjpiicuF7cz+7yeZycia2GULue4evIbCCNobDtU
GHungAECgYEA6JOjuOKHnMWMEgc5oyHtW7BXr18Z8XhQ8eyxKuYptPrdyUw9agte
JtL7nDTuL1wwc7URjBB2ymuwpAuhb5vxrD360UrUJWEavtnSZK3cZYa1WxL0I2y0
D3qQgeObGFy91J3EwA9KGV5PNsPf9rJDdMJgcvJqZlTpqzb23zIHgmECgYEAxIzs
zOmZz8pL0gmQN6+okMNc6kjvDwNnYquh5ZvDxo/7XoDY0VInwtc+ZUttTA21w91N
fLbZKbRA2Ee1F7cGoV7u1V4YocT6at1GAlnZxfV3QIqcPTNREONYWX0aPd+b+y9k
4dcYNAJdCYN/dqNMWQQYtC70GPTAVR3xaBiP6gECgYEA5Xtv62iReM2vJMa+R0md
o3+/NUo4Ffuqmtr6ASMzeeCiYBH68xyeXN6G552Oe2qSYEkENFi2bYqOs37KXo7X
iiVpy4LzCqLiuffBUhf+xKqDXYa5IA8NJ8y+s3r6OLKhmB3H2d38NkXJEXd6EDfa
uWVlt2WcOLaGDathMd9ya2ECgYByMJmm1xS4ZwQzy8CQyan6KLZDmwngRA79gU92
sU9FfgMBPYQ54Cwfg6PJf8/I/rIaT+kjyqtSEloWDVsFoxzkBd5l8dwHqAQAr/tr
hD4ER3737U+mMrknQZ3jp83mIpJhlYBbwPZbyP+6dj5Ic8j4cmvTyu+fzBotmU7W
NmbuAQKBgFyyL/ONPSyUU7JjlvRYO2Pw56bRVUac6Li0d1k0BpUVllOfHx9et/+t
MGnZhrCkQjLRWw8PrVd9mp2UzDbthEfuvJeaPETeGxJH7mZpz2+n1+Iz8x7bTAZz
tjS/+eO3udolhm1FAXtz0I+4sZA1u0QzO+ScYXBrlt+0yGMlSqm/
-----END RSA PRIVATE KEY-----
EOF
    }
  }
}

resource "aws_key_pair" "deployer" {
  key_name   = "deployer-key-%[1]s"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCykRE5W1q7Yc5OBH7lMRh23uiGKK+IoxRGUXOATq3YRGlGGNTEboOel2fShfL/0EVT3BaXnGr5U0xZefgYIcAg6VtvlmjsT9buPC88A5915jUosKOOWbSrZFnNGG4cwhAsU/pZVyM49CXJT5DiXtdAWVRyli7OQLSIImFJF8lDot7tzeReCebVaPAaIaKVA2WdCytrTNNdzuNMia/irFpEZQq2nhtDab80cBkWVy2Obnpm85xZS7Aa/u3Uc6XnyF2B/aRKLKTWHqyIFh2jfNwbKi6j5FPIR5mvHFMsbghxApAFruWC/xPX109SXTYAnNpN6IsdBTeE7BsD6Wz7Yyxh"
}
`, randId)
}
