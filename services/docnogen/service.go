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
	GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error)
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

func (s *docnogenService) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error) {
	// check if Repository has been initialized

	if s.DocNoRepo == nil || s.DocNoFormatter == nil {
		out = &pb.GenerateBulkDocNoFormatResponse{
			Ok:           false,
			ErrorCode:    500,
			ErrorMessage: fmt.Sprint("Document Number Repository or Document Number Formatter is nil"),
			Results:      []*pb.GenerateBulkDocNoFormatResponse_Result{},
		}
	} else {
		var preCondiErr error
		// check if DocCode is empty
		if in.DocCode == "" {
			preCondiErr = fmt.Errorf("Doc Code is empty")
		}

		// check if OrgCode is empty
		if in.OrgCode == "" {
			preCondiErr = fmt.Errorf("Organisation Code is empty")
		}

		// check if Path is empty
		if in.Path == "" {
			preCondiErr = fmt.Errorf("Path is empty")
		}

		// check if BulkNumber is at least 1 and not more than 99
		if in.BulkNumber < 1 || in.BulkNumber > 99 {
			preCondiErr = fmt.Errorf("Bulk Number must be at least 1 and not more than 99")
		}

		// check if Format string is empty
		format := s.getFormatString(in.OrgCode, in.DocCode, in.Path, in.CustomFormat)
		if format == "" {
			preCondiErr = fmt.Errorf("Format is empty")
		}

		// if no error for preconditions
		if preCondiErr == nil {
			// start generating document for x number of times (based on BulkNumber)
			var docNo *models.DocNo
			var updatedDoc *models.DocNo
			results := []*pb.GenerateBulkDocNoFormatResponse_Result{}
			for x := 0; x < int(in.BulkNumber); x++ {
				fmt.Printf("trying to generate doc number for %d/%d...\n", x+1, in.BulkNumber)
				// try get and consume the sequence number until successful, else if error because concurrency update detechted... keep trying
				updateSuccess := false
				for {
					// this loop will loop infinite until updateSuccess = true
					docNo, err = s.DocNoRepo.GetByPath(in.DocCode, in.OrgCode, in.Path)
					if err != nil {
						out = &pb.GenerateBulkDocNoFormatResponse{
							Ok:           false,
							ErrorCode:    500,
							ErrorMessage: err.Error(),
							Results:      results,
						}

						break
					} else {
						// if no error, and document not nil, assign result to response
						if docNo != nil {
							// generate Sequence Number string
							seqNoStr := s.DocNoFormatter.GenerateSeqNoStr(in.OrgCode, in.DocCode, in.Path, docNo.NextSeqNo)
							if seqNoStr == "" {
								err = fmt.Errorf("Sequence Number String is empty")
								out = &pb.GenerateBulkDocNoFormatResponse{
									Ok:           false,
									ErrorCode:    400,
									ErrorMessage: err.Error(),
									Results:      results,
								}

								break
							}

							// generate Document Number string
							var docNoStr string
							docNoStr, err = s.DocNoFormatter.GenerateFormatString(format, in.DocCode, seqNoStr, in.VariableMap)
							if err != nil {
								out = &pb.GenerateBulkDocNoFormatResponse{
									Ok:           false,
									ErrorCode:    400,
									ErrorMessage: err.Error(),
									Results:      results,
								}

								break
							}

							var r pb.GenerateBulkDocNoFormatResponse_Result

							// consume the sequence number
							// increase sequence number by 1 and set a new record timestamp to mark record has been altered
							currSeqNo := docNo.NextSeqNo
							currRecordTimestamp := docNo.RecordTimestamp
							docNo.NextSeqNo++
							docNo.RecordTimestamp = time.Now().Unix()
							// update the doc to db with concurrency update control
							updatedDoc, err = s.DocNoRepo.UpdateByPath(in.OrgCode, docNo, currSeqNo, currRecordTimestamp)
							if err != nil && err != common.ConcurrencyUpdateError {
								out = &pb.GenerateBulkDocNoFormatResponse{
									Ok:           false,
									ErrorCode:    500,
									ErrorMessage: err.Error(),
									Results:      results,
								}

								break
							} else if err != nil && err == common.ConcurrencyUpdateError {
								// concurrency update error detected...
								// loop again
							} else {
								// no error, update successful
								if updatedDoc != nil {
									r = pb.GenerateBulkDocNoFormatResponse_Result{
										DocNoString:     docNoStr,
										NextSeqNo:       uint32(updatedDoc.NextSeqNo),
										RecordTimestamp: updatedDoc.RecordTimestamp,
									}
								}
								// add r to results
								results = append(results, &r)

								updateSuccess = true
							}
						} else {
							// else if no document found (which should never hanppen because system will auto generate initial document with sequence number start with 1)
							out = &pb.GenerateBulkDocNoFormatResponse{
								Ok:           false,
								ErrorCode:    500,
								ErrorMessage: "Something is wrong with the system...",
								Results:      results,
							}
							break
						}
					}

					if updateSuccess {
						// if consume success, break from the loop
						fmt.Println(results)
						break
					}
				}

			}
			// end of loop

			fmt.Printf("out=%v err=%v", out, err)
			if err == nil {
				// genereate OK response
				out = &pb.GenerateBulkDocNoFormatResponse{
					Ok:           true,
					ErrorCode:    0,
					ErrorMessage: "",
					Results:      results,
				}
			}

		} else {
			// preconditions have errors
			out = &pb.GenerateBulkDocNoFormatResponse{
				Ok:           false,
				ErrorCode:    400,
				ErrorMessage: preCondiErr.Error(),
				Results:      []*pb.GenerateBulkDocNoFormatResponse_Result{},
			}
		}
	}

	return out, nil
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

		// check if OrgCode is empty
		if in.OrgCode == "" {
			preCondiErr = fmt.Errorf("Organisation Code is empty")
		}

		// check if Path is empty
		if in.Path == "" {
			preCondiErr = fmt.Errorf("Path is empty")
		}

		// check if Format string is empty
		format := s.getFormatString(in.OrgCode, in.DocCode, in.Path, in.CustomFormat)
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
								ErrorCode:    400,
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
								ErrorCode:    400,
								ErrorMessage: err.Error(),
								Result:       nil,
							}

							break
						}

						var result pb.GenerateDocNoFormatResponse_Result

						// consume the sequence number
						// increase sequence number by 1 and set a new record timestamp to mark record has been altered
						currSeqNo := docNo.NextSeqNo
						currRecordTimestamp := docNo.RecordTimestamp
						docNo.NextSeqNo++
						docNo.RecordTimestamp = time.Now().Unix()
						// update the doc to db with concurrency update control
						updatedDoc, err := s.DocNoRepo.UpdateByPath(in.OrgCode, docNo, currSeqNo, currRecordTimestamp)
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
				ErrorCode:    400,
				ErrorMessage: preCondiErr.Error(),
				Result:       nil,
			}
		}
	}

	return out, nil
}

func (s *docnogenService) GetNextDocNo(ctx context.Context, in *pb.GetNextDocNoRequest) (out *pb.GetNextDocNoResponse, err error) {
	if s.DocNoRepo == nil || s.DocNoFormatter == nil {
		out = &pb.GetNextDocNoResponse{
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

		// check if OrgCode is empty
		if in.OrgCode == "" {
			preCondiErr = fmt.Errorf("Organisation Code is empty")
		}

		// check if Path is empty
		if in.Path == "" {
			preCondiErr = fmt.Errorf("Path is empty")
		}

		// check if Format string is empty
		format := s.getFormatString(in.OrgCode, in.DocCode, in.Path, in.CustomFormat)
		if format == "" {
			preCondiErr = fmt.Errorf("Format is empty")
		}

		// if no error for preconditions
		if preCondiErr == nil {
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
					var docNoStr string
					// generate Sequence Number string
					seqNoStr := s.DocNoFormatter.GenerateSeqNoStr(in.OrgCode, in.DocCode, in.Path, docNo.NextSeqNo)
					if seqNoStr == "" {
						err = fmt.Errorf("Sequence Number String is empty")
					} else {
						// generate Document Number string
						docNoStr, err = s.DocNoFormatter.GenerateFormatString(format, in.DocCode, seqNoStr, in.VariableMap)
						fmt.Printf("docNoStr=%s err=%v\n", docNoStr, err)

					}

					if err != nil {
						out = &pb.GetNextDocNoResponse{
							Ok:           false,
							ErrorCode:    400,
							ErrorMessage: err.Error(),
							Result:       nil,
						}
					} else {

						result = pb.GetNextDocNoResponse_Result{
							DocNoString:     docNoStr,
							NextSeqNo:       uint32(docNo.NextSeqNo),
							RecordTimestamp: docNo.RecordTimestamp,
						}

						out = &pb.GetNextDocNoResponse{
							Ok:           true,
							ErrorCode:    0,
							ErrorMessage: "",
							Result:       &result,
						}
					}
				}
			}
		} else {
			// preconditions have errors
			out = &pb.GetNextDocNoResponse{
				Ok:           false,
				ErrorCode:    400,
				ErrorMessage: preCondiErr.Error(),
				Result:       nil,
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
		var preCondiErr error
		// check if DocCode is empty
		if in.DocCode == "" {
			preCondiErr = fmt.Errorf("Doc Code is empty")
		}

		// check if OrgCode is empty
		if in.OrgCode == "" {
			preCondiErr = fmt.Errorf("Organisation Code is empty")
		}

		// check if Path is empty
		if in.Path == "" {
			preCondiErr = fmt.Errorf("Path is empty")
		}

		// if no error for preconditions
		if preCondiErr == nil {
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
					updatedDoc, err := s.DocNoRepo.UpdateByPath(in.OrgCode, docNo, int64(in.CurSeqNo), currRecordTimestamp)
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
		} else {
			// preconditions have errors
			out = &pb.ConsumeDocNoResponse{
				Ok:           false,
				ErrorCode:    400,
				ErrorMessage: preCondiErr.Error(),
				Result:       nil,
			}
		}
	}

	return out, nil
}

// This internal function check if Custom Function is passed in from request, if yes, Custom Function will be return
func (s *docnogenService) getFormatString(orgCode string, docCode string, path string, customFormat string) string {
	if customFormat != "" {
		fmt.Println("Custom Format is defined")
		return customFormat
	}

	fmt.Printf("Custom Format is not defined, system format is generated according to parameters: OrgCode=%s DocCode=%s Path=%s\n", orgCode, docCode, path)
	return s.DocNoFormatter.GetFormatString(orgCode, docCode, path)
}
