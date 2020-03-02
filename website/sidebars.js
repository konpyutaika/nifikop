/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

module.exports = {
    "docs": 
    {
      "Overview": ["overview"],
      "Casskop" : 
        [
          {
            "type" : "category", 
            "label": "Getting Started", 
            "items"  : [
              "casskop/1_getting_started/1_overview", 
              "casskop/1_getting_started/2_Pre-requisites", 
              "casskop/1_getting_started/3_quickstart"
            ]
          },
          { 
            "type" : "category", 
            "label": "Deployment Configuration",
            "items"  : [
              "casskop/2_deployment_configuration/1_cassandra_cluster_config",
              "casskop/2_deployment_configuration/2_cassandra_config",
              "casskop/2_deployment_configuration/3_cassandra_storage",
              "casskop/2_deployment_configuration/4_kubernetest_object",
              "casskop/2_deployment_configuration/5_cpu_memory_resources",
              "casskop/2_deployment_configuration/6_cluster_topology",
              "casskop/2_deployment_configuration/7_implementation_architecture",
              "casskop/2_deployment_configuration/8_advanced_configuration",
              "casskop/2_deployment_configuration/9_cassandra_node_management",
              "casskop/2_deployment_configuration/10_cassandracluster_status",
              "casskop/2_deployment_configuration/11_cassandracluster_crd_definition"
            ]
          },
          { 
            "type" : "category", 
            "label": "Operations",
            "items"  : [
              "casskop/3_operations/1_overview",
              "casskop/3_operations/2_cluster_operations",
              "casskop/3_operations/3_cassandra_pods_operations"
            ]
          },
          "casskop/4_troubleshooting"
        ],
      "Multi-Casskop":
        [
          "multi-casskop/1_overview",
          "multi-casskop/2_pre-requisite",
          "multi-casskop/3_quickstart"
        ],
      "Contributing" : 
        [
          {
            "type" : "category", 
            "label": "Development", 
            "items"  : [
              "contributing/1_development/1_circle_ci_build_pipeline",
              "contributing/1_development/2_operator_sdk"
            ]
          },
          {
            "type" : "category", 
            "label": "Release the project", 
            "items"  : [
              "contributing/2_release_project/1_tasks",
              "contributing/2_release_project/2_with_helm",
              "contributing/2_release_project/3_with_olm"
            ]
          },
          "contributing/3_reporting_bugs",
          {
            "type" : "category", 
            "label": "How this repository was initially build", 
            "items"  : [
              "contributing/4_initial_build/1_boilerplate_casskop",
              "contributing/4_initial_build/2_infos_developer"
            ]
          }
        ]
    }
    
};
