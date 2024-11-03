package main

import (
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

func CleanQueue(sonarr *sonarr.Sonarr) error {
	err := sonarr.DeleteQueue(1, &starr.QueueDeleteOpts{})
	return err
}
