package api

var _ saveState = (*saveStateBroken)(nil)

type saveStateBroken struct {
	saveStateCommon
}
