package api

var _ saveState = (*saveStateFinished)(nil)

type saveStateFinished struct {
	saveStateCommon
}
