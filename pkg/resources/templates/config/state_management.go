package config

var StateManagementTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<stateManagement>
    <local-provider>
        <id>local-provider</id>
        <class>org.apache.nifi.controller.state.providers.local.WriteAheadLocalStateProvider</class>
        <property name="Directory">../data/state/local</property>
        <property name="Always Sync">false</property>
        <property name="Partitions">16</property>
        <property name="Checkpoint Interval">2 mins</property>
    </local-provider>
    <cluster-provider>
        <id>zk-provider</id>
        <class>org.apache.nifi.controller.state.providers.zookeeper.ZooKeeperStateProvider</class>
        <property name="Connect String">{{ .ZookeeperConnectString }}</property>
        <property name="Root Node">{{ .ZookeeperPath }}</property>
        <property name="Session Timeout">10 seconds</property>
        <property name="Access Control">Open</property>
    </cluster-provider>    
    <cluster-provider>
        <id>kubernetes-provider</id>
        <class>org.apache.nifi.kubernetes.state.provider.KubernetesConfigMapStateProvider</class>
    </cluster-provider>
</stateManagement>
`
