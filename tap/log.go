package tap

import "github.com/sirupsen/logrus"

func LogError(err error) {
	if err != nil {
		logrus.Error(err)
	}
}
