/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"errors"
	"net/rpc"
	"sync"

	"github.com/cgrates/cgrates/utils"
)

type StatsInterface interface {
	AddQueue(*StatsQueue, *int) error
	GetValues(string, *map[string]float64) error
	AppendCDR(*utils.StoredCdr, *int) error
}

type Stats struct {
	queues map[string]*StatsQueue
	mux    sync.RWMutex
}

func (s *Stats) AddQueue(sq *StatsQueue, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.queues[sq.conf.Id] = sq
	*out = 0
	return nil
}

func (s *Stats) GetValues(sqID string, values *map[string]float64) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if sq, ok := s.queues[sqID]; ok {
		*values = sq.GetStats()
		return nil
	}
	return errors.New("Not Found")
}

func (s *Stats) AppendCDR(cdr *utils.StoredCdr, out *int) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, sq := range s.queues {
		sq.AppendCDR(cdr)
	}
	*out = 0
	return nil
}

type ProxyStats struct {
	Client *rpc.Client
}

func NewProxyStats(addr string) (*ProxyStats, error) {
	client, err := rpc.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}
	return &ProxyStats{Client: client}, nil
}

func (ps *ProxyStats) AddQueue(sq *StatsQueue, out *int) error {
	return ps.Client.Call("Scribe.AddQueue", sq, out)
}

func (ps *ProxyStats) GetValues(sqID string, values *map[string]float64) error {
	return ps.Client.Call("Scribe.GetValues", sqID, values)
}

func (ps *ProxyStats) AppendCDR(cdr *utils.StoredCdr, out *int) error {
	return ps.Client.Call("Scribe.AppendCDR", cdr, out)
}