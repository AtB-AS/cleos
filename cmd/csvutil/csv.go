package main

import "encoding/csv"

type csvReaderImpl struct {
	r      *csv.Reader
	header []string
}

func (c *csvReaderImpl) readHeader() error {
	hdr, err := c.r.Read()
	if err != nil {
		return err
	}

	c.header = hdr
	return nil
}

func (c *csvReaderImpl) Header() ([]string, error) {
	if len(c.header) < 1 {
		if err := c.readHeader(); err != nil {
			return nil, err
		}
	}
	return c.header, nil
}

func (c *csvReaderImpl) Next() ([]string, error) {
	row, err := c.r.Read()
	if err != nil {
		return nil, err
	}

	return row, nil
}
