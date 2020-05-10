package database

// const (
// 	etcdDialTimeout    = 5 * time.Second
// 	etcdRequestTimeout = 500 * time.Millisecond
// )

// type EtcdDatabase struct {
// 	client clientv3.KV
// 	conn   *clientv3.Client
// }

// func NewEtcdDatabase(endpoints []string) (*EtcdDatabase, error) {
// 	e := &EtcdDatabase{}

// 	var err error
// 	e.conn, err = clientv3.New(clientv3.Config{
// 		Endpoints:   endpoints,
// 		DialTimeout: etcdDialTimeout,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	e.client = e.conn.KV

// 	return e, nil
// }

// // func (d *EtcdDatabase) AddPrefix(prefix string) datastore.Database {
// // 	return &EtcdDatabase{
// // 		client: namespace.NewKV(d.client, prefix+"/"),
// // 		conn:   d.conn,
// // 	}
// // }

// func defaultKey(key string) string {
// 	return "default/" + key
// }
// func deletedKey(key string) string {
// 	return "deleted/" + key
// }

// func (d EtcdDatabase) List(ctx context.Context, f func(int) []proto.Message) error {
// 	c, cancel := context.WithTimeout(ctx, etcdRequestTimeout)
// 	defer cancel()

// 	resp, err := d.client.Get(c, defaultKey(""), clientv3.WithFromKey())
// 	if err != nil {
// 		return errors.Wrap(err, "Get()")
// 	}
// 	if len(resp.Kvs) == 0 {
// 		return ErrorNotFound
// 	}

// 	pb := f(len(resp.Kvs))

// 	for i, ev := range resp.Kvs {
// 		err = proto.Unmarshal(ev.Value, pb[i])
// 		if err != nil {
// 			return errors.Wrap(err, "proto.Unmarshal()")
// 		}
// 	}

// 	return nil
// }

// func (d EtcdDatabase) Get(ctx context.Context, name string, entity Entity) error {
// 	resp, err := d.client.Get(ctx, defaultKey(name))
// 	if err != nil {
// 		return err
// 	}
// 	if resp.Count == 0 {
// 		entity = nil
// 		return ErrorNotFound
// 	}

// 	err = proto.Unmarshal(resp.Kvs[0].Value, entity)
// 	if err != nil {
// 		return errors.Wrap(err, "proto.Unmarshal()")
// 	}

// 	return nil
// }

// func (d EtcdDatabase) Apply(ctx context.Context, entity Entity) error {
// 	md := entity.GetMetadata()

// 	s, err := proto.Marshal(entity)
// 	if err != nil {
// 		return errors.Wrapf(err, "proto.Marshal(%v)", entity)
// 	}

// 	txn := d.client.Txn(ctx)
// 	txnRes, err := txn.
// 		If(clientv3.Compare(clientv3.Version(defaultKey(md.Name)), "=", md.Revision)).
// 		Then(clientv3.OpPut(defaultKey(md.Name), string(s))).
// 		Commit()
// 	if err != nil {
// 		return err
// 	}

// 	if !txnRes.Succeeded {
// 		return ErrorConflict
// 	}

// 	return nil
// }

// func (d EtcdDatabase) SoftDelete(ctx context.Context, entity Entity) error {
// 	md := entity.GetMetadata()
// 	md.DeletedAt = ptypes.TimestampNow()

// 	s, err := proto.Marshal(entity)
// 	if err != nil {
// 		return errors.Wrapf(err, "proto.Marshal(%v)", entity)
// 	}

// 	txn := d.client.Txn(ctx)
// 	txnRes, err := txn.
// 		If(clientv3.Compare(clientv3.Version(md.Name), "=", md.Revision)).
// 		Then(
// 			clientv3.OpDelete(defaultKey(md.Name)),
// 			clientv3.OpPut(deletedKey(md.Uid), string(s)),
// 		).
// 		Commit()
// 	if err != nil {
// 		return errors.Wrap(err, "txn.Commit()")
// 	}

// 	if !txnRes.Succeeded {
// 		return ErrorConflict
// 	}

// 	return nil
// }

// func (d EtcdDatabase) HardDelete(ctx context.Context, entity Entity) error {
// 	return fmt.Errorf("unimplemented")
// }

// func (d EtcdDatabase) Close() error {
// 	return d.conn.Close()
// }
