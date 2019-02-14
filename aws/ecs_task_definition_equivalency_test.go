package aws

import (
	"testing"
)

func TestAwsEcsContainerDefinitionsAreEquivalent_basic(t *testing.T) {
	cfgRepresention := `
[
    {
      "name": "wordpress",
      "links": [
        "mysql"
      ],
      "image": "wordpress",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 80,
          "hostPort": 80
        }
      ],
      "memory": 500,
      "cpu": 10
    },
    {
      "environment": [
        {
          "name": "MYSQL_ROOT_PASSWORD",
          "value": "password"
        }
      ],
      "name": "mysql",
      "image": "mysql",
      "cpu": 10,
      "memory": 500,
      "essential": true
    }
]`

	apiRepresentation := `
[
    {
        "name": "wordpress",
        "image": "wordpress",
        "cpu": 10,
        "memory": 500,
        "links": [
            "mysql"
        ],
        "portMappings": [
            {
                "containerPort": 80,
                "hostPort": 80,
                "protocol": "tcp"
            }
        ],
        "essential": true,
        "environment": [],
        "mountPoints": [],
        "volumesFrom": []
    },
    {
        "name": "mysql",
        "image": "mysql",
        "cpu": 10,
        "memory": 500,
        "portMappings": [],
        "essential": true,
        "environment": [
            {
                "name": "MYSQL_ROOT_PASSWORD",
                "value": "password"
            }
        ],
        "mountPoints": [],
        "volumesFrom": []
    }
]`

	equal, err := EcsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestAwsEcsContainerDefinitionsAreEquivalent_portMappings(t *testing.T) {
	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 80
        }
      ],
      "memory": 500,
      "cpu": 10
    }
]`

	apiRepresentation := `
[
    {
        "name": "wordpress",
        "image": "wordpress",
        "cpu": 10,
        "memory": 500,
        "portMappings": [
            {
                "containerPort": 80,
                "hostPort": 0,
                "protocol": "tcp"
            }
        ],
        "essential": true,
        "environment": [],
        "mountPoints": [],
        "volumesFrom": []
    }
]`

	equal, err := EcsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestAwsEcsContainerDefinitionsAreEquivalent_portMappingsIgnoreHostPort(t *testing.T) {
	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "portMappings": [
        {
          "containerPort": 80,
          "hostPort": 80
        }
      ]
    }
]`

	apiRepresentation := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "portMappings": [
        {
          "containerPort": 80
        }
      ]
    }
]`

	var (
		equal bool
		err   error
	)

	equal, err = EcsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Fatal("Expected definitions to differ.")
	}

	equal, err = EcsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, true)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestAwsEcsContainerDefinitionsAreEquivalent_arrays(t *testing.T) {
	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "links": ["container1", "container2", "container3"],
      "portMappings": [
        {"containerPort": 80},
        {"containerPort": 81},
        {"containerPort": 82}
      ],
      "environment": [
        {"name": "VARNAME1", "value": "VARVAL1"},
        {"name": "VARNAME2", "value": "VARVAL2"},
        {"name": "VARNAME3", "value": "VARVAL3"}
      ],
      "extraHosts": [
        {"hostname": "host1", "ipAddress": "127.0.0.1"},
        {"hostname": "host2", "ipAddress": "127.0.0.2"},
        {"hostname": "host3", "ipAddress": "127.0.0.3"}
      ],
      "mountPoints": [
        {"sourceVolume": "vol1", "containerPath": "/vol1"},
        {"sourceVolume": "vol2", "containerPath": "/vol2"},
        {"sourceVolume": "vol3", "containerPath": "/vol3"}
      ],
      "volumesFrom": [
        {"sourceContainer": "container1"},
        {"sourceContainer": "container2"},
        {"sourceContainer": "container3"}
      ],
      "ulimits": [
        {
          "name": "core",
          "softLimit": 10, "hardLimit": 20
        },
        {
          "name": "cpu",
          "softLimit": 10, "hardLimit": 20
        },
        {
          "name": "fsize",
          "softLimit": 10, "hardLimit": 20
        }
      ],
      "linuxParameters": {
        "capabilities": {
          "add": ["AUDIT_CONTROL", "AUDIT_WRITE", "BLOCK_SUSPEND"],
          "drop": ["CHOWN", "IPC_LOCK", "KILL"]
        }
      },
      "devices": [
        {
          "hostPath": "/path1",
          "permissions": ["read", "write", "mknod"]
        },
        {
          "hostPath": "/path2",
          "permissions": ["read", "write"]
        },
        {
          "hostPath": "/path3",
          "permissions": ["read", "mknod"]
        }
      ],
      "dockerSecurityOptions": ["label:one", "label:two", "label:three"],
      "memory": 500,
      "cpu": 10
    },
    {
      "name": "container1",
      "image": "busybox",
      "memory": 100
    },
    {
      "name": "container2",
      "image": "busybox",
      "memory": 100
    },
    {
      "name": "container3",
      "image": "busybox",
      "memory": 100
    }
]`

	apiRepresentation := `
[
  {
    "cpu": 10,
    "dockerSecurityOptions": [
      "label:one",
      "label:two",
      "label:three"
    ],
    "environment": [
      {
        "name": "VARNAME3",
        "value": "VARVAL3"
      },
      {
        "name": "VARNAME2",
        "value": "VARVAL2"
      },
      {
        "name": "VARNAME1",
        "value": "VARVAL1"
      }
    ],
    "essential": true,
    "extraHosts": [
      {
        "hostname": "host1",
        "ipAddress": "127.0.0.1"
      },
      {
        "hostname": "host2",
        "ipAddress": "127.0.0.2"
      },
      {
        "hostname": "host3",
        "ipAddress": "127.0.0.3"
      }
    ],
    "image": "wordpress",
    "links": [
      "container1",
      "container2",
      "container3"
    ],
    "linuxParameters": {
      "capabilities": {
        "add": [
          "AUDIT_CONTROL",
          "AUDIT_WRITE",
          "BLOCK_SUSPEND"
        ],
        "drop": [
          "CHOWN",
          "IPC_LOCK",
          "KILL"
        ]
      }
    },
    "memory": 500,
    "mountPoints": [
      {
        "containerPath": "/vol1",
        "sourceVolume": "vol1"
      },
      {
        "containerPath": "/vol2",
        "sourceVolume": "vol2"
      },
      {
        "containerPath": "/vol3",
        "sourceVolume": "vol3"
      }
    ],
    "name": "wordpress",
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 0,
        "protocol": "tcp"
      },
      {
        "containerPort": 81,
        "hostPort": 0,
        "protocol": "tcp"
      },
      {
        "containerPort": 82,
        "hostPort": 0,
        "protocol": "tcp"
      }
    ],
    "ulimits": [
      {
        "hardLimit": 20,
        "name": "core",
        "softLimit": 10
      },
      {
        "hardLimit": 20,
        "name": "cpu",
        "softLimit": 10
      },
      {
        "hardLimit": 20,
        "name": "fsize",
        "softLimit": 10
      }
    ],
    "volumesFrom": [
      {
        "sourceContainer": "container1"
      },
      {
        "sourceContainer": "container2"
      },
      {
        "sourceContainer": "container3"
      }
    ]
  },
  {
    "cpu": 0,
    "environment": [],
    "essential": true,
    "image": "busybox",
    "memory": 100,
    "mountPoints": [],
    "name": "container1",
    "portMappings": [],
    "volumesFrom": []
  },
  {
    "cpu": 0,
    "environment": [],
    "essential": true,
    "image": "busybox",
    "memory": 100,
    "mountPoints": [],
    "name": "container2",
    "portMappings": [],
    "volumesFrom": []
  },
  {
    "cpu": 0,
    "environment": [],
    "essential": true,
    "image": "busybox",
    "memory": 100,
    "mountPoints": [],
    "name": "container3",
    "portMappings": [],
    "volumesFrom": []
  }
]
`

	equal, err := EcsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestAwsEcsContainerDefinitionsAreEquivalent_negative(t *testing.T) {
	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "memory": 500,
      "cpu": 10,
      "environment": [
        {"name": "EXAMPLE_NAME", "value": "foobar"}
      ]
    }
]`

	apiRepresentation := `
[
    {
        "name": "wordpress",
        "image": "wordpress",
        "cpu": 10,
        "memory": 500,
        "essential": true,
        "environment": [],
        "mountPoints": [],
        "volumesFrom": []
    }
]`

	equal, err := EcsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Fatal("Expected definitions to differ.")
	}
}
