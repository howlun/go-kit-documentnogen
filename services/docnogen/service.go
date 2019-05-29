package docnogensvc

import (
	"fmt"
	"time"

	pb "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/pb"
	context "golang.org/x/net/context"

	"github.com/howlun/go-kit-documentnogen/common"
	"github.com/howlun/go-kit-documentnogen/services/docnogen/models"
)

type DocnogenService interface {
	GenerateDocNoFormat(ctx context.Context, in *pb.GenerateDocNoFormatRequest) (out *pb.GenerateDocNoFormatResponse, err error)
	GetNextDocNo(ctx context.Context, in *pb.GetNextDocNoRequest) (out *pb.GetNextDocNoResponse, err error)
	ConsumeDocNo(ctx context.Context, in *pb.ConsumeDocNoRequest) (out *pb.ConsumeDocNoResponse, err error)
}

type docnogenService struct {
	DocNoRepo      models.DocNoRepository
	DocNoFormatter DocnoformatterService
}

func NewDocnogenService(repo models.DocNoRepository, formatter DocnoformatterService) (s pb.DocNoGenServiceServer) {
	s = &docnogenService{DocNoRepo: repo, DocNoFormatter: formatter}
	return s
}

func (s *docnogenService) GenerateDocNoFormat(ctx context.Context, in *pb.GenerateDocNoFormatRequest) (out *pb.GenerateDocNoFormatResponse, err error) {
	// check if Repository has been initialized
	if s.DocNoRepo == nil || s.DocNoFormatter == nil {
		out = &pb.GenerateDocNoFormatResponse{
			Ok:           false,
			ErrorCode:    500,
			ErrorMessage: fmt.Sprint("Document Number Repository or Document Number Formatter is nil"),
			Result:       nil,
		}
	} else {
		var preCondiErr error
		// check if DocCode is empty
		if in.DocCode == "" {
			preCondiErr = fmt.Errorf("Doc Code is empty")
		}

		// check if Format string is empty
		format := s.DocNoFormatter.GetFormatString(in.OrgCode, in.DocCode, in.Path)
		if format == "" {
			preCondiErr = fmt.Errorf("Format is empty")
		}

		// if no error for preconditions
		if preCondiErr == nil {
			// try get and consume the sequence number until successful, else if error because concurrency update detechted... keep trying
			updateSuccess := false
			for {
				// this loop will loop infinite until updateSuccess = true
				docNo, err := s.DocNoRepo.GetByPath(in.DocCode, in.OrgCode, in.Path)
				if err != nil {
					out = &pb.GenerateDocNoFormatResponse{
						Ok:           false,
						ErrorCode:    500,
						ErrorMessage: err.Error(),
						Result:       nil,
					}

					break
				} else {
					// if no error, and document not nil, assign result to response
					if docNo != nil {
						// generate Sequence Number string
						seqNoStr := s.DocNoFormatter.GenerateSeqNoStr(in.OrgCode, in.DocCode, in.Path, docNo.NextSeqNo)
						if seqNoStr == "" {
							out = &pb.GenerateDocNoFormatResponse{
								Ok:           false,
								ErrorCode:    500,
								ErrorMessage: "Sequence Number String is empty",
								Result:       nil,
							}

							break
						}

						// generate Document Number string
						docNoStr, err := s.DocNoFormatter.GenerateFormatString(format, in.DocCode, seqNoStr, in.VariableMap)
						if err != nil {
							out = &pb.GenerateDocNoFormatResponse{
								Ok:           false,
								ErrorCode:    500,
								ErrorMessage: err.Error(),
								Result:       nil,
							}

							break
						}

						var result pb.GenerateDocNoFormatResponse_Result

						// consume the sequence number
						// increase sequence number by 1 and set a new record timestamp to mark record has been altered
						currRecordTimestamp := docNo.RecordTimestamp
						docNo.NextSeqNo++
						docNo.RecordTimestamp = time.Now().Unix()
						// update the doc to db with concurrency update control
						updatedDoc, err := s.DocNoRepo.UpdateByPath(in.OrgCode, docNo, currRecordTimestamp)
						if err != nil && err != common.ConcurrencyUpdateError {
							out = &pb.GenerateDocNoFormatResponse{
								Ok:           false,
								ErrorCode:    500,
								ErrorMessage: err.Error(),
								Result:       nil,
							}

							break
						} else if err != nil && err == common.ConcurrencyUpdateError {
							// concurrency update error detected...
							// loop again
						} else {
							// no error, update successful
							if updatedDoc != nil {
								result = pb.GenerateDocNoFormatResponse_Result{
									DocNoString:     docNoStr,
									NextSeqNo:       uint32(updatedDoc.NextSeqNo),
									RecordTimestamp: updatedDoc.RecordTimestamp,
								}
							}

							out = &pb.GenerateDocNoFormatResponse{
								Ok:           true,
								ErrorCode:    0,
								ErrorMessage: "",
								Result:       &result,
							}

							updateSuccess = true
						}
					} else {
						// else if no document found (which should never hanppen because system will auto generate initial document with sequence number start with 1)
						out = &pb.GenerateDocNoFormatResponse{
							Ok:           false,
							ErrorCode:    500,
							ErrorMessage: "Something is wrong with the system...",
							Result:       nil,
						}
						break
					}
				}

				if updateSuccess {
					// if consume success, break from the loop
					break
				}
			}
		} else {
			// preconditions have errors
			out = &pb.GenerateDocNoFormatResponse{
				Ok:           false,
				ErrorCode:    500,
				ErrorMessage: err.Error(),
				Result:       nil,
			}
		}
	}

	return out, nil
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
				currRecordTimestamp := docNo.RecordTimestamp
				docNo.NextSeqNo++
				docNo.RecordTimestamp = time.Now().Unix()
				// update the doc to db with concurrency update control
				updatedDoc, err := s.DocNoRepo.UpdateByPath(in.OrgCode, docNo, currRecordTimestamp)
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
