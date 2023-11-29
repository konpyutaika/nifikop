package reportingtask

import (
	"strconv"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("reportingtask-method")

const (
	reportingTaskName                = "managed-prometheus"
	reportingTaskType_               = "org.apache.nifi.reporting.prometheus.PrometheusReportingTask"
	reportingTaskEnpointPortProperty = "prometheus-reporting-task-metrics-endpoint-port"
	reportingTaskStrategyProperty    = "prometheus-reporting-task-metrics-strategy"
	reportingTaskStrategy            = "All Components"
	reportingTaskSendJVMProperty     = "prometheus-reporting-task-metrics-send-jvm"
	reportingTaskSendJVM             = "true"
)

func ExistReportingTaks(config *clientconfig.NifiConfig, cluster *v1.NifiCluster) (bool, error) {
	if cluster.Status.PrometheusReportingTask.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetReportingTask(cluster.Status.PrometheusReportingTask.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get reporting-task"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

func CreateReportingTask(config *clientconfig.NifiConfig, cluster *v1.NifiCluster) (*v1.PrometheusReportingTaskStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ReportingTaskEntity{}
	updateReportingTaskEntity(cluster, &scratchEntity)

	entity, err := nClient.CreateReportingTask(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create reporting-task"); err != nil {
		return nil, err
	}

	return &v1.PrometheusReportingTaskStatus{
		Id:      entity.Id,
		Version: *entity.Revision.Version,
	}, nil
}

func SyncReportingTask(config *clientconfig.NifiConfig, cluster *v1.NifiCluster) (*v1.PrometheusReportingTaskStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetReportingTask(cluster.Status.PrometheusReportingTask.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get registry-client"); err != nil {
		return nil, err
	}

	if !reportingTaksIsSync(cluster, entity) {
		status := entity.Status

		if status.ValidationStatus == "VALIDATING" {
			return nil, errorfactory.NifiReportingTasksValidating{}
		}

		if status.RunStatus == "RUNNING" {
			entity, err = nClient.UpdateRunStatusReportingTask(entity.Id, nigoapi.ReportingTaskRunStatusEntity{
				Revision: entity.Revision,
				State:    "STOPPED",
			})
			if err := clientwrappers.ErrorUpdateOperation(log, err, "Update reporting-task status"); err != nil {
				return nil, err
			}
		}

		updateReportingTaskEntity(cluster, entity)
		entity, err = nClient.UpdateReportingTask(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update reporting-task"); err != nil {
			return nil, err
		}
	}

	if entity.Status.ValidationStatus == "INVALID" {
		return nil, errorfactory.NifiReportingTasksInvalid{}
	}

	if entity.Status.RunStatus == "STOPPED" || entity.Status.RunStatus == "DISABLED" {
		log.Info("Starting Prometheus reporting task",
			zap.String("clusterName", cluster.Name))
		entity, err = nClient.UpdateRunStatusReportingTask(entity.Id, nigoapi.ReportingTaskRunStatusEntity{
			Revision: entity.Revision,
			State:    "RUNNING",
		})
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update reporting-task status"); err != nil {
			return nil, err
		}
	}

	status := cluster.Status.PrometheusReportingTask
	status.Version = *entity.Revision.Version
	status.Id = entity.Id

	return &status, nil
}

func RemoveReportingTaks(config *clientconfig.NifiConfig, cluster *v1.NifiCluster) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetReportingTask(cluster.Status.PrometheusReportingTask.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get reporting-task"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	updateReportingTaskEntity(cluster, entity)
	err = nClient.RemoveReportingTask(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove registry-client")
}

func reportingTaksIsSync(cluster *v1.NifiCluster, entity *nigoapi.ReportingTaskEntity) bool {
	return reportingTaskName == entity.Component.Name &&
		strconv.Itoa(*cluster.Spec.GetMetricPort()) == entity.Component.Properties[reportingTaskEnpointPortProperty] &&
		reportingTaskStrategy == entity.Component.Properties[reportingTaskStrategyProperty] &&
		reportingTaskSendJVM == entity.Component.Properties[reportingTaskSendJVMProperty]
}

func updateReportingTaskEntity(cluster *v1.NifiCluster, entity *nigoapi.ReportingTaskEntity) {
	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.ReportingTaskEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.ReportingTaskDto{}
	}

	entity.Component.Name = "managed-prometheus"
	entity.Component.Type_ = "org.apache.nifi.reporting.prometheus.PrometheusReportingTask"
	entity.Component.Properties = map[string]string{
		reportingTaskEnpointPortProperty: strconv.Itoa(*cluster.Spec.GetMetricPort()),
		reportingTaskStrategyProperty:    reportingTaskStrategy,
		reportingTaskSendJVMProperty:     reportingTaskSendJVM,
	}
}
