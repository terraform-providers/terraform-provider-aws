package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSGlueScript_Language_Python(t *testing.T) {
	dataSourceName := "data.aws_glue_script.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSGlueScriptConfigPython(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "python_script", regexp.MustCompile(`from awsglue\.job import Job`)),
				),
			},
		},
	})
}

func TestAccDataSourceAWSGlueScript_Language_Scala(t *testing.T) {
	dataSourceName := "data.aws_glue_script.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSGlueScriptConfigScala(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "scala_code", regexp.MustCompile(`import com\.amazonaws\.services\.glue\.util\.Job`)),
				),
			},
		},
	})
}

func testAccDataSourceAWSGlueScriptConfigPython() string {
	return `
data "aws_glue_script" "test" {
  language = "PYTHON"

  dag_edge {
    source = "datasource0"
    target = "applymapping1"
  }

  dag_edge {
    source = "applymapping1"
    target = "selectfields2"
  }

  dag_edge {
    source = "selectfields2"
    target = "resolvechoice3"
  }

  dag_edge {
    source = "resolvechoice3"
    target = "datasink4"
  }

  dag_node {
    id        = "datasource0"
    node_type = "DataSource"

    args {
      name  = "database"
      value = "\"SourceDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"SourceTable\""
    }
  }

  dag_node {
    id        = "applymapping1"
    node_type = "ApplyMapping"

    args {
      name  = "mapping"
      value = "[(\"column1\", \"string\", \"column1\", \"string\")]"
    }
  }

  dag_node {
    id        = "selectfields2"
    node_type = "SelectFields"

    args {
      name  = "paths"
      value = "[\"column1\"]"
    }
  }

  dag_node {
    id        = "resolvechoice3"
    node_type = "ResolveChoice"

    args {
      name  = "choice"
      value = "\"MATCH_CATALOG\""
    }

    args {
      name  = "database"
      value = "\"DestinationDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"DestinationTable\""
    }
  }

  dag_node {
    id        = "datasink4"
    node_type = "DataSink"

    args {
      name  = "database"
      value = "\"DestinationDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"DestinationTable\""
    }
  }
}
`
}

func testAccDataSourceAWSGlueScriptConfigScala() string {
	return `
data "aws_glue_script" "test" {
  language = "SCALA"

  dag_edge {
    source = "datasource0"
    target = "applymapping1"
  }

  dag_edge {
    source = "applymapping1"
    target = "selectfields2"
  }

  dag_edge {
    source = "selectfields2"
    target = "resolvechoice3"
  }

  dag_edge {
    source = "resolvechoice3"
    target = "datasink4"
  }

  dag_node {
    id        = "datasource0"
    node_type = "DataSource"

    args {
      name  = "database"
      value = "\"SourceDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"SourceTable\""
    }
  }

  dag_node {
    id        = "applymapping1"
    node_type = "ApplyMapping"

    args {
      name  = "mappings"
      value = "[(\"column1\", \"string\", \"column1\", \"string\")]"
    }
  }

  dag_node {
    id        = "selectfields2"
    node_type = "SelectFields"

    args {
      name  = "paths"
      value = "[\"column1\"]"
    }
  }

  dag_node {
    id        = "resolvechoice3"
    node_type = "ResolveChoice"

    args {
      name  = "choice"
      value = "\"MATCH_CATALOG\""
    }

    args {
      name  = "database"
      value = "\"DestinationDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"DestinationTable\""
    }
  }

  dag_node {
    id        = "datasink4"
    node_type = "DataSink"

    args {
      name  = "database"
      value = "\"DestinationDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"DestinationTable\""
    }
  }
}
`
}
