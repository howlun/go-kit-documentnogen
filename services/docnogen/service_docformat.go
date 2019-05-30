package docnogensvc

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/howlun/go-kit-documentnogen/common"
)

type DocnoformatterService interface {
	GetFormatString(orgCode string, docCode string, path string) string
	GenerateSeqNoStr(orgCode string, docCode string, path string, seqNo int64) string
	SplitFormatToArray(format string) []string
	ValidateFormatString(format string, docCode string, seqNoStr string, variableMap map[string]string) (bool, error)
	GenerateFormatString(format string, docCode string, seqNoStr string, variableMap map[string]string) (string, error)
}

type docNoFormatterDefaultService struct {
}

func NewDocnoformatterService() (s DocnoformatterService) {
	s = &docNoFormatterDefaultService{}
	return s
}

func (df *docNoFormatterDefaultService) GetFormatString(orgCode string, docCode string, path string) string {
	return common.DefaultDocFormat
}

func (df *docNoFormatterDefaultService) GenerateSeqNoStr(orgCode string, docCode string, path string, seqNo int64) string {
	return fmt.Sprintf(common.DefaultSeqNoFormat, common.DefaultSeqNoLength, seqNo)
}

func (df *docNoFormatterDefaultService) SplitFormatToArray(format string) []string {
	re := regexp.MustCompile(common.MustCompilePatternStr)
	arr := re.FindAllString(format, -1)

	for i, str := range arr {
		str = strings.Replace(str, "{{", "", -1)
		str = strings.Replace(str, "}}", "", -1)
		arr[i] = str
	}
	fmt.Println(arr)
	return arr
}

// This function check if all the variables in the Variable Map able to map to the Format required
func (df *docNoFormatterDefaultService) ValidateFormatString(format string, docCode string, seqNoStr string, variableMap map[string]string) (bool, error) {
	validateSuccess := true
	hasFixedVarPrefix := false
	hasFixedVarSeqNo := false
	var err error

	// addin two more fixed Variable: PREFIX and SEQNO
	variableMap[common.FixedVarPrefix] = docCode
	variableMap[common.FixedVarSeqNo] = seqNoStr

	for _, varName := range df.SplitFormatToArray(format) {
		// check if the mandatory variables (PREFIX and SEQNO) are provided
		if varName == common.FixedVarPrefix {
			hasFixedVarPrefix = true
		} else if varName == common.FixedVarSeqNo {
			hasFixedVarSeqNo = true
		}

		if _, ok := variableMap[varName]; !ok {
			validateSuccess = false
			err = fmt.Errorf("The value for one of the custom variables defined in Format is not provided or not found: {{%s}}", varName)

			break
		}
	}

	if !hasFixedVarPrefix || !hasFixedVarSeqNo {
		validateSuccess = false
		err = fmt.Errorf("The required variable {{%s}} and/or {{%s}} in Format is not provided or not found", common.FixedVarPrefix, common.FixedVarSeqNo)
	}

	fmt.Printf("Format is valid=%v err=%v\n", validateSuccess, err)
	return validateSuccess, err
}

func (df *docNoFormatterDefaultService) GenerateFormatString(format string, docCode string, seqNoStr string, variableMap map[string]string) (string, error) {
	if format == "" {
		return "", fmt.Errorf("Format string is empty")
	}

	if docCode == "" {
		return "", fmt.Errorf("Doc Code is empty")
	}

	if seqNoStr == "" {
		return "", fmt.Errorf("Sequence Number String is empty")
	}

	// Check if all required variables needed in Format is provided in variable Map
	fmt.Println("Check if all required variables needed in Format is provided in variable Map")
	formatIsValid, err := df.ValidateFormatString(format, docCode, seqNoStr, variableMap)
	if formatIsValid == false {
		return "", fmt.Errorf("Format is not valid with Variable Map: %s", err.Error())
	}

	// addin two more fixed Variable: PREFIX and SEQNO
	variableMap[common.FixedVarPrefix] = docCode
	variableMap[common.FixedVarSeqNo] = seqNoStr

	//initialize docNoString with format
	docNoString := format
	// Replace variable into Format string
	for _, key := range df.SplitFormatToArray(format) {
		docNoString = strings.Replace(docNoString, fmt.Sprintf("{{%s}}", key), variableMap[key], -1)
	}
	fmt.Printf("docNoString=%s err=%v\n", docNoString, err)
	return docNoString, nil
}
