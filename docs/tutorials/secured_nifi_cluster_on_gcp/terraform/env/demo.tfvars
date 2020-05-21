# GCP configurations
project  = "poc-rtc"
region   = "europe-west1"
zone     = "europe-west1-c"

# GKE Cluster
cluster_machines_types = "n1-standard-2"

# GKE variables
username = "demo"
password = "demodemodemodemo"
min_node = 1
max_node = 6
initial_node_count = 5
preemptible = false

# NiFiKop configuration
## Image
nifikop_image_repo = "eu.gcr.io/poc-rtc/nifikop"
nifikop_image_tag  = "0.0.1-feature_tls_certmanager"
nifikop_chart_version = "0.0.1-feature-http-client"