package app

import (
	"testing"

	l "github.com/sanek1/metrics-collector/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	wantLogger := &l.ZapLogger{}

	gotLogger, gotError := InitLogger()

	assert.Equal(t, wantLogger, gotLogger)
	assert.Equal(t, nil, gotError)
}
