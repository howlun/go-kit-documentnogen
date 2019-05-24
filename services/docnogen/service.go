package docnogensvc

import (
	"fmt"
	"time"

	pb "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/pb"
	context "golang.org/x/net/context"

	"github.com/howlun/go-kit-documentnogen/services/docnogen/models"
)

type DocnogenService interface {
	GetNextDocNo(ctx context.Context, in *pb.GetNextDocNoRequest) (out *pb.GetNextDocNoResponse, err error)
	ConsumeDocNo(ctx context.Context, in *pb.ConsumeDocNoRequest) (out *pb.ConsumeDocNoResponse, err error)
}

type docnogenService struct {
	DocNoRepo models.DocNoRepository
}

func NewDocnogenService(repo models.DocNoRepository) (s pb.DocNoGenServiceServer) {
	// TODO: Implement initialization of service
	s = &docnogenService{DocNoRepo: repo}
	return s
}

func (s *docnogenService) GetNextDocNo(ctx context.Context, in *pb.GetNextDocNoRequest) (out *pb.GetNextDocNoResponse, err error) {
	// check if Repository has been initialized
	if s.DocNoRepo == nil {
		out = &pb.GetNextDocNoResponse{
			Ok:           false,
			ErrorCode:    500,
			ErrorMessage: fmt.Sprint("Document Number Repository is nil"),
			Result:       nil,
		}
	} else {

		// Call GetByPath to get document
		var result pb.GetNextDocNoResponse_Result
		docNo, err := s.DocNoRepo.GetByPath(in.DocCode, in.OrgCode, in.Path)
		if err != nil {
			out = &pb.GetNextDocNoResponse{
				Ok:           false,
				ErrorCode:    500,
				ErrorMessage: err.Error(),
				Result:       nil,
			}
		} else {
			// if no error, and document not nil, assign result to response
			if docNo != nil {
				result = pb.GetNextDocNoResponse_Result{
					NextSeqNo:       uint32(docNo.NextSeqNo),
					RecordTimestamp: docNo.RecordTimestamp,
				}
			}
			out = &pb.GetNextDocNoResponse{
				Ok:           true,
				ErrorCode:    0,
				ErrorMessage: "",
				Result:       &result,
			}
		}
	}
	return out, nil
}

func (s *docnogenService) ConsumeDocNo(ctx context.Context, in *pb.ConsumeDocNoRequest) (out *pb.ConsumeDocNoResponse, err error) {
	// check if Repository has been initialized
	if s.DocNoRepo == nil {
		out = &pb.ConsumeDocNoResponse{
			Ok:           false,
			ErrorCode:    500,
			ErrorMessage: fmt.Sprint("Document Number Repository is nil"),
			Result:       nil,
		}
	} else {

		// Call GetByPath to get document
		var result pb.ConsumeDocNoResponse_Result
		docNo, err := s.DocNoRepo.GetByPath(in.DocCode, in.OrgCode, in.Path)
		if err != nil {
			out = &pb.ConsumeDocNoResponse{
				Ok:           false,
				ErrorCode:    500,
				ErrorMessage: err.Error(),
				Result:       nil,
			}
		} else {
			// if no error, and document not found, throw error document not found
			if docNo == nil {
				out = &pb.ConsumeDocNoResponse{
					Ok:           false,
					ErrorCode:    500,
					ErrorMessage: fmt.Sprintf("No document found with OrgCode=%s DocCode=%s Path=%s", in.OrgCode, in.DocCode, in.Path),
					Result:       nil,
				}
				//err = fmt.Errorf("No document found with OrgCode=%s DocCode=%s Path=%s", in.OrgCode, in.DocCode, in.Path)
			} else if uint32(docNo.NextSeqNo) != in.CurSeqNo {
				// compare record sequence number, if not the same, record has been altered and throw error concurrency update
				out = &pb.ConsumeDocNoResponse{
					Ok:           false,
					ErrorCode:    400,
					ErrorMessage: fmt.Sprintf("Concurreny Update error with OrgCode=%s DocCode=%s Path=%s UserSeqNumber=%d SystemSeqNumber=%d", in.OrgCode, in.DocCode, in.Path, in.CurSeqNo, docNo.NextSeqNo),
					Result:       nil,
				}
				//err = fmt.Errorf("Concurreny Update error with OrgCode=%s DocCode=%s Path=%s UserSeqNumber=%d SystemSeqNumber=%d", in.OrgCode, in.DocCode, in.Path, in.CurSeqNo, docNo.NextSeqNo)
			} else if docNo.RecordTimestamp != in.RecordTimestamp {
				// compare record timestamp, if not the same, record has been altered and throw error concurrency update
				out = &pb.ConsumeDocNoResponse{
					Ok:           false,
					ErrorCode:    400,
					ErrorMessage: fmt.Sprintf("Concurreny Update error with OrgCode=%s DocCode=%s Path=%s UserSubmitted=%d System=%d", in.OrgCode, in.DocCode, in.Path, in.RecordTimestamp, docNo.RecordTimestamp),
					Result:       nil,
				}
				//err = fmt.Errorf("Concurreny Update error with OrgCode=%s DocCode=%s Path=%s UserSubmitted=%d System=%d", in.OrgCode, in.DocCode, in.Path, in.RecordTimestamp, docNo.RecordTimestamp)
			} else {
				// increase sequence number by 1 and set a new record timestamp to mark record has been altered
				docNo.NextSeqNo++
				docNo.RecordTimestamp = time.Now().Unix()
				// update the doc to db
				updatedDoc, err := s.DocNoRepo.UpdateByPath(in.OrgCode, docNo)
				if err != nil {
					out = &pb.ConsumeDocNoResponse{
						Ok:           false,
						ErrorCode:    500,
						ErrorMessage: err.Error(),
						Result:       nil,
					}
				} else {
					if updatedDoc != nil {
						result = pb.ConsumeDocNoResponse_Result{
							NextSeqNo:       uint32(updatedDoc.NextSeqNo),
							RecordTimestamp: updatedDoc.RecordTimestamp,
						}
					}

					out = &pb.ConsumeDocNoResponse{
						Ok:           true,
						ErrorCode:    0,
						ErrorMessage: "",
						Result:       &result,
					}
				}
			}

		}
	}

	return out, nil
}
