/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

module.exports = {
    "docs":
        {
            "Concepts": [
                "1_concepts/1_introduction",
                "1_concepts/2_design_principes",
                "1_concepts/3_features",
                "1_concepts/4_roadmap",
            ],
            "Setup": [
                "2_setup/1_getting_started",
                {
                    "type" : "category",
                    "label": "Platform Setup",
                    "items"  : [
                        "2_setup/2_platform_setup/1_gke",
                        "2_setup/2_platform_setup/2_minikube",
//                    "2_setup/2_platform_setup/3_microk8s",
//                   "2_setup/2_platform_setup/4_docker_desktop",
                    ]
                },
                {
                    "type" : "category",
                    "label": "Install",
                    "items"  : [
                        "2_setup/3_install/1_customizable_install_with_helm",
                    ]
                }
            ],
            "Tasks": [
                {
                    "type" : "category",
                    "label": "NiFi Cluster",
                    "items"  : [
//                        "3_tasks/1_nifi_cluster/1_nodes_configuration",
                        "3_tasks/1_nifi_cluster/2_cluster_scaling",
//                        "3_tasks/1_nifi_cluster/3_external_dns",

                    ]
                },
                {
                    "type" : "category",
                    "label": "Security",
                    "items"  : [
                        "3_tasks/2_security/1_ssl",
                        {
                            "type" : "category",
                            "label": "Authentication",
                            "items"  : [
                                "3_tasks/2_security/2_authentication/1_oidc",
                            ]
                        },
                    ]
                },
                "3_tasks/3_nifi_dataflow",
                "3_tasks/4_nifi_user_group"
            ],
            // "Examples": [
            //     "4_examples/1_simple_nifi_cluster"
            // ],
            "Reference": [
                {
                    "type" : "category",
                    "label": "NiFi Cluster",
                    "items"  : [
                        "5_references/1_nifi_cluster/1_nifi_cluster",
                        "5_references/1_nifi_cluster/2_read_only_config",
                        "5_references/1_nifi_cluster/3_node_config",
                        "5_references/1_nifi_cluster/4_node",
                        "5_references/1_nifi_cluster/5_node_state",
                        "5_references/1_nifi_cluster/6_listeners_config",
                    ]
                },
                "5_references/2_nifi_user",
                "5_references/3_nifi_registry_client",
                "5_references/4_nifi_parameter_context",
                "5_references/5_nifi_dataflow",
                "5_references/6_nifi_usergroup",

            ],
            "Contributing" : [
                "6_contributing/1_developer_guide",
                "6_contributing/2_reporting_bugs",
                "6_contributing/3_credits",
            ],
        }
};