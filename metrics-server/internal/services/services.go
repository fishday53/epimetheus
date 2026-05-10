package services

import (
	ctx "context"
	pb "metrics-server/internal/proto"
	"metrics-server/internal/usecase"
	"metrics-server/internal/usecase/context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricServiceServer struct {
	pb.UnimplementedMetricServiceServer
	app *context.AppContext
}

func NewMetricServiceServer(app *context.AppContext) *MetricServiceServer {
	return &MetricServiceServer{
		app: app,
	}
}

// SetMultiParamGRPC serves incoming metric batches.
func (s *MetricServiceServer) SetMultiParam(ctx ctx.Context, in *pb.SetMultiParamRequest) (*pb.SetMultiParamResponse, error) {
	var response pb.SetMultiParamResponse

	mBatch := convertFromProtoBatch(in.Metrics)

	for _, metric := range mBatch {
		if metric.ID == "" {
			s.app.Log.Errorln("Name is not defined")
			return &response, status.Error(codes.InvalidArgument, "Name is not defined")
		}
		_, err := s.app.DB.Set(metric)
		if err != nil {
			s.app.Log.Errorln("Cannot set metric:", err)
			return &response, status.Error(codes.Internal, err.Error())
		}
	}

	if s.app.Cfg.StoreInterval == 0 {
		err := s.app.DB.Dump(s.app.Cfg.FileStoragePath)
		if err != nil {
			s.app.Log.Errorln("Dump error:", err)
			return &response, status.Error(codes.Internal, err.Error())
		}
	}

	s.app.Log.Infof("Got %d metrics via GRPC.", len(mBatch))
	return &response, nil
}

// GetParamGRPC sends stored metric.
func (s *MetricServiceServer) GetParam(ctx ctx.Context, in *pb.GetParamRequest) (*pb.GetParamResponse, error) {
	var (
		response pb.GetParamResponse
		metric   *usecase.Metric
	)

	metric = convertFromProtoMetric(in.Metric)

	if metric.ID == "" {
		s.app.Log.Errorln("Name is not defined")
		return &response, status.Error(codes.InvalidArgument, "Name is not defined")
	}

	result, err := s.app.DB.Get(metric)
	if err != nil {
		s.app.Log.Errorln("Cannot get metric:", err)
		return &response, status.Error(codes.InvalidArgument, "Not found")
	}

	response.Metric = convertToProtoMetric(result)

	return &response, nil
}

// GetParamGRPC sends all stored metrics.
func (s *MetricServiceServer) GetAllParams(ctx ctx.Context, in *pb.GetAllParamsRequest) (*pb.GetAllParamsResponse, error) {
	var response pb.GetAllParamsResponse

	result, err := s.app.DB.GetAll()
	if err != nil {
		s.app.Log.Errorln("Cannot get all metrics:", err)
		return &response, status.Error(codes.Internal, err.Error())
	}

	response.Metrics = convertToProtoBatch(result)

	return &response, nil
}

func convertFromProtoMetric(pm *pb.Metric) *usecase.Metric {
	if pm == nil {
		return nil
	}

	return &usecase.Metric{
		ID:    pm.Id,
		MType: pm.Mtype,
		Delta: &pm.Delta,
		Value: &pm.Value,
	}
}

func convertFromProtoBatch(batch *pb.Batch) []*usecase.Metric {
	if batch == nil {
		return nil
	}

	mBatch := make([]*usecase.Metric, 0, len(batch.Metrics))

	for _, m := range batch.Metrics {
		mBatch = append(mBatch, convertFromProtoMetric(m))
	}

	return mBatch
}

func convertToProtoMetric(m *usecase.Metric) *pb.Metric {
	return &pb.Metric{
		Id:    m.ID,
		Mtype: m.MType,
		Delta: *m.Delta,
		Value: *m.Value,
	}
}

func convertToProtoBatch(batch *[]usecase.Metric) *pb.Batch {
	if batch == nil {
		return nil
	}

	pbBatch := &pb.Batch{
		Metrics: make([]*pb.Metric, 0, len(*batch)),
	}

	for _, m := range *batch {
		pbBatch.Metrics = append(pbBatch.Metrics, convertToProtoMetric(&m))
	}

	return pbBatch
}
