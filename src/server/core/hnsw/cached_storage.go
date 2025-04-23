package hnsw

type Buffered struct {
	workspace int64
	buffer    map[int64]*nodeData
	deleted   map[int64]struct{}

	originalEntryPoint *int64
	entryPoint         *int64
}

type BufferedStorage struct {
	storage *Storage
	buffers map[int64]*Buffered
}

func NewBufferedStorage(storage *Storage) *BufferedStorage {
	return &BufferedStorage{
		storage: storage,
		buffers: make(map[int64]*Buffered),
	}
}

func (s *BufferedStorage) GetBuffered(workspace int64) *Buffered {
	b, ok := s.buffers[workspace]
	if !ok {
		b = &Buffered{
			workspace: workspace,
			buffer:    make(map[int64]*nodeData),
			deleted:   make(map[int64]struct{}),
		}
		s.buffers[workspace] = b
	}

	return b
}

func (s *BufferedStorage) SaveNode(workspace int64, id int64, data nodeData) error {
	b := s.GetBuffered(workspace)
	b.buffer[id] = &data
	delete(b.deleted, id)
	return nil
}

func (s *BufferedStorage) DeleteNode(workspace int64, id int64) error {
	b := s.GetBuffered(workspace)

	delete(b.buffer, id)
	b.deleted[id] = struct{}{}
	return nil
}

func (s *BufferedStorage) LoadNode(workspace int64, id int64) (*nodeData, error) {
	b := s.GetBuffered(workspace)
	data, ok := b.buffer[id]
	if ok {
		return data, nil
	}

	if _, ok := b.deleted[id]; ok {
		return nil, ErrNotFound
	}

	d, err := s.storage.LoadNode(workspace, id)
	if err != nil {
		return nil, err
	}

	b.buffer[id] = &d

	return &d, nil
}

func (s *BufferedStorage) SetEntryPoint(workspace int64, id int64) error {
	b := s.GetBuffered(workspace)
	b.originalEntryPoint = b.entryPoint
	b.entryPoint = &id
	return nil
}

func (s *BufferedStorage) Flush() error {
	for workspace, b := range s.buffers {
		for id, data := range b.buffer {
			if data.IsDirty() {
				if err := s.storage.SaveNode(workspace, id, *data); err != nil {
					return err
				}
				data.dirty = false
			}
		}

		for id := range b.deleted {
			if err := s.storage.DeleteNode(workspace, id); err != nil {
				return err
			}
		}

		if b.entryPoint != nil {
			if err := s.storage.SetEntryPoint(workspace, *b.entryPoint); err != nil {
				return err
			}
		}

		b.buffer = make(map[int64]*nodeData)
		b.deleted = make(map[int64]struct{})
	}
	return nil
}

func (s *BufferedStorage) Close() error {
	if err := s.Flush(); err != nil {
		return err
	}
	return s.storage.Close()
}
