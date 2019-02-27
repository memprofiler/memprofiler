package backend

var _ saveState = (*saveStateFinished)(nil)

type saveStateFinished struct {
	saveStateCommon
}
