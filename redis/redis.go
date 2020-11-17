package redis

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/keimoon/gore"
)

var (
	ErrNotOk    = errors.New("not ok")
	ErrBadReply = errors.New("bad reply")
	ErrNoPool   = errors.New("no connection pool")
)

type Redis struct {
	ConnURL string
	Pool    *gore.Pool
}

// Creates a new connection via URL string and authenticates it for you
func NewRedis(connURL string) (*Redis, error) {
	u, err := url.Parse(connURL)
	if err != nil {
		return nil, err
	}
	pass, _ := u.User.Password()
	pool := &gore.Pool{
		RequestTimeout: time.Second * 1,
		Password:       pass,
	}
	err = pool.Dial(fmt.Sprintf("%s:%s", u.Hostname(), u.Port()))
	if err != nil {
		return nil, err
	}
	return &Redis{
		Pool:    pool,
		ConnURL: connURL,
	}, err
}

func (r *Redis) Close() {
	if r.Pool == nil {
		return
	}
	r.Pool.Close()
}

func (r *Redis) RunCmd(command string, args ...interface{}) (*gore.Reply, error) {
	if r.Pool == nil {
		return nil, ErrNoPool
	}
	conn, err := r.Pool.Acquire()
	if err != nil {
		return nil, err
	}
	defer r.Pool.Release(conn)
	if r.Pool.Password != "" {
		if err := conn.Auth(r.Pool.Password); err != nil {
			return nil, err
		}
	}
	reply, err := gore.NewCommand(command, args...).Run(conn)
	if err != nil {
		return nil, err
	}
	if reply.IsError() {
		errMsg, _ := reply.Error()
		return nil, errors.New(errMsg)
	}
	return reply, nil
}

func (r *Redis) Scan(cursor int, match string) (xs []string, newCursor int, err error) {
	var reply *gore.Reply
	reply, err = r.RunCmd("SCAN", cursor, "MATCH", match)
	if err != nil {
		return
	}

	var arr []*gore.Reply
	arr, err = reply.Array()
	if err != nil {
		return
	} else if len(arr) != 2 {
		err = ErrBadReply
		return
	}

	if c, err := arr[0].Int(); err == nil {
		newCursor = int(c)
	}

	var rr []*gore.Reply
	rr, err = arr[1].Array()
	if err != nil {
		return
	}
	for _, x := range rr {
		if str, err := x.String(); err == nil {
			xs = append(xs, str)
		}
	}
	return
}

// Yes, all keys in Redis are strings
func (r *Redis) Keys(match string) ([]string, error) {
	reply, err := r.RunCmd("KEYS", match)
	if err != nil {
		return nil, err
	}
	rs, err := reply.Array()
	// In the case where we didn't get an array type we just assume no keys
	if err != nil {
		return []string{}, nil
	}

	var xs []string
	for _, x := range rs {
		if str, err := x.String(); err != nil {
			xs = append(xs, str)
		}
	}
	return xs, nil
}

func (r *Redis) Del(keys ...string) error {
	var xs []interface{}
	for _, x := range keys {
		xs = append(xs, x)
	}
	_, err := r.RunCmd("DEL", xs...)
	if err != nil {
		return err
	}
	return nil
}

// NOTICE: https://redis.io/commands/set
// Since the SET command options can replace SETNX, SETEX, PSETEX, GETSET, it is possible that in future versions of
// Redis these three commands will be deprecated and finally removed.
//
// Hence, let's not implement them. - Tom
func (r *Redis) Set(key string, value interface{}, expire time.Duration) error {
	args := []interface{}{key, value}
	if expire > 0 {
		args = append(args, "EX", int(expire.Seconds()))
	}

	reply, err := r.RunCmd("SET", args...)
	if err != nil {
		return err
	} else if !reply.IsOk() {
		return ErrNotOk
	}
	return nil
}

func (r *Redis) Get(key string) ([]byte, bool, error) {
	reply, err := r.RunCmd("GET", key)
	if err != nil {
		return nil, false, err
	}
	b, err := reply.Bytes()
	if err == gore.ErrNil {
		return b, false, nil
	}
	return b, err == nil, err
}

func (r *Redis) DeleteKeyMatch(match string) (int, error) {
	return r.DeleteKeyMatchFn(match, nil)
}

func (r *Redis) ScanAll(match string) ([]string, error) {
	allKeys := make([]string, 0, 100)
	var cursor int
	fn := func() error {
		var keys []string
		var err error
		keys, cursor, err = r.Scan(cursor, match)
		if err != nil {
			return err
		}
		allKeys = append(allKeys, keys...)
		return nil
	}

	err := fn()
	for cursor != 0 && err == nil {
		err = fn()
	}
	return allKeys, err
}

type DeleteErrors struct {
	Errors map[string]error
}

func (e *DeleteErrors) Error() string {
	return fmt.Sprintf("Could not delete all keys")
}
func (e *DeleteErrors) Add(key string, err error) {
	e.Errors[key] = err
}

func (r *Redis) DeleteKeyMatchFn(match string, keyFn func(string)) (int, error) {
	delErrs := &DeleteErrors{}

	var cursor, count int
	fn := func() error {
		var keys []string
		var err error
		keys, cursor, err = r.Scan(cursor, match)
		if err != nil {
			return err
		}
		for _, key := range keys {
			count++
			if err := r.Del(key); err != nil {
				delErrs.Add(key, err)
			} else if keyFn != nil {
				keyFn(key)
			}
		}
		return nil
	}

	err := fn()
	for cursor != 0 && err == nil {
		err = fn()
	}
	if err == nil && len(delErrs.Errors) > 0 {
		err = delErrs
	}
	if err == nil {
		// Ensure that the [type] of the [nil] error variables is <nil,nil> and not <error,nil>
		return count, nil
	}
	return count, err
}
