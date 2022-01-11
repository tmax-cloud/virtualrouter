package vpn

import (
	"github.com/strongswan/govici/vici"
)

func loadConn(conn Connection) error {
	s, err := vici.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()

	c, err := vici.MarshalMessage(&conn)
	if err != nil {
		return err
	}

	m := vici.NewMessage()
	if err := m.Set(conn.Name, c); err != nil {
		return err
	}

	_, err = s.CommandRequest("load-conn", m)

	return err
}

func unloadConn(name string) error {
	s, err := vici.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()

	type Conn struct {
		Name string `vici:"name"`
	}

	conn := Conn{
		Name: name,
	}

	m, err := vici.MarshalMessage(&conn)
	if err != nil {
		return err
	}

	_, err = s.CommandRequest("unload-conn", m)

	return err
}

func getConns() ([]string, error) {
	s, err := vici.NewSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	m, err := s.CommandRequest("get-conns", nil)
	if err != nil {
		return nil, err
	}

	return m.Keys(), nil
}

func loadShared(secret Secret) error {
	s, err := vici.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()

	m, err := vici.MarshalMessage(&secret)
	if err != nil {
		return err
	}

	_, err = s.CommandRequest("load-shared", m)

	return err
}

func unloadShared(id string) error {
	s, err := vici.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()

	secret := Secret{
		ID: id,
	}

	m, err := vici.MarshalMessage(&secret)
	if err != nil {
		return err
	}

	_, err = s.CommandRequest("unload-shared", m)

	return err
}

func getShared() ([]string, error) {
	s, err := vici.NewSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	m, err := s.CommandRequest("get-shared", nil)
	if err != nil {
		return nil, err
	}

	return m.Keys(), nil
}
