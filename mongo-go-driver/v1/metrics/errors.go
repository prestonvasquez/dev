package metrics

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// TooManyLogicalSessions will return "true" if the error is
// "TooManyLogicalSessions"
func ErrorIsTooManyLogicalSessions(err error) bool {
	if err != nil {
		srvErr, ok := err.(mongo.ServerError)
		if ok {
			//errSet[err.Error()]++
			return srvErr.HasErrorCode(261)
		}
	}

	return false
}
