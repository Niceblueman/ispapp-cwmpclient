package cron

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Cron struct {
	*cron.Cron
	logger cron.Logger
	jobs   []*cron.EntryID
}

func New() *Cron {
	// check if development mode
	logger := cron.PrintfLogger(logrus.StandardLogger())
	_cron := &Cron{
		logger: logger,
		Cron:   cron.New(cron.WithChain(cron.Recover(logger)), cron.WithLocation(time.UTC), cron.WithLogger(logger)),
		jobs:   make([]*cron.EntryID, 0),
	}
	return _cron
}

func (c *Cron) AddJob(spec string, cmd cron.Job) {
	id, err := c.Cron.AddJob(spec, cmd)
	if err != nil {
		c.logger.Error(err, "failed to add cron job")
		return
	}
	c.jobs = append(c.jobs, &id)
	c.logger.Info("added cron job: ", spec)
}
