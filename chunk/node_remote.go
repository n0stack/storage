package chunk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	echo "github.com/labstack/echo"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

const DEFAULT_POST_CHUNK_TIMEOUT = 200 * time.Microsecond

const (
	PARAM_CHUNK_ID            = "chunk_id"
	WRITE_QUERY_PRE_CHECKSUM  = "pre_checksum"
	WRITE_QUERY_POST_CHECKSUM = "post_checksum"
	WRITE_QUERY_OFFSET        = "offset"
	WRITE_QUERY_REPLICA_NODES = "replica_nodes"
	WRITE_FORM_DATA           = "data"
)

type RemoteChunkNode struct {
	Timeout time.Duration
}

func OpenRemoteChunkNode() *RemoteChunkNode {
	return &RemoteChunkNode{
		Timeout: DEFAULT_POST_CHUNK_TIMEOUT,
	}
}

// req.Nodesの扱いに注意
func (n *RemoteChunkNode) WriteChunk(ctx context.Context, req *WriteChunkRequest) (*WriteChunkResponse, error) {
	res := &WriteChunkResponse{
		HealthNodes: make([]string, 0),
		DesyncNodes: make([]string, 0),
	}
	bodyBuf := &bytes.Buffer{}
	body := multipart.NewWriter(bodyBuf)

	// TODO: set file name
	fw, err := body.CreateFormFile(WRITE_FORM_DATA, WRITE_FORM_DATA)
	if err != nil {
		return nil, errors.Wrap(err, "CreateFormFile()")
	}

	if _, err := io.Copy(fw, req.Data); err != nil {
		return nil, errors.Wrap(err, "io.Copy()")
	}

	contentType := body.FormDataContentType()
	body.Close()

	for i, node := range req.Nodes {
		values := url.Values{}
		values.Add(WRITE_QUERY_PRE_CHECKSUM, fmt.Sprintf("%016x", req.PreChecksum))
		values.Add(WRITE_QUERY_POST_CHECKSUM, fmt.Sprintf("%016x", req.PostChecksum))
		values.Add(WRITE_QUERY_OFFSET, fmt.Sprintf("%d", req.Offset))
		values.Add(WRITE_QUERY_REPLICA_NODES, strings.Join(req.Nodes[i:], ","))

		client := http.DefaultClient
		req, err := http.NewRequest("POST", filepath.Join(node, req.ID.String())+"?"+values.Encode(), bodyBuf)
		if err != nil {
			log.Printf("[WARNING] failed to struct request to %s, err=%v", node, err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), n.Timeout)
		defer cancel()
		req = req.WithContext(ctx)

		req.Header.Add("Content-Type", contentType)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[WARNING] failed to replicate to %s, err=%v", node, err)
			res.DesyncNodes = append(res.DesyncNodes, node)
			continue
		} else {
			defer resp.Body.Close()

			// TODO: error handling

			r := &WriteChunkResponse{}
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("[WARNING] invalid response from %s, err=%v", node, err)
				continue
			}

			json.Unmarshal(b, r)
			res.HealthNodes = append(res.HealthNodes, r.HealthNodes...)
			res.DesyncNodes = append(res.DesyncNodes, r.DesyncNodes...)

			break
		}
	}

	return res, nil
}

type ChunkStoreService struct {
	root         string
	MaxChunkSize int64
}

func NewChunkStoreService(root string) *ChunkStoreService {
	return &ChunkStoreService{
		root:         root,
		MaxChunkSize: DEFAULT_MAX_CHUNK_SIZE,
	}
}

func (s ChunkStoreService) Filepath(id uuid.UUID) string {
	return filepath.Join(s.root, id.String())
}

func (s ChunkStoreService) ReadChunk(ectx echo.Context) error {
	chunkID := ectx.Param(PARAM_CHUNK_ID)
	id, err := uuid.FromString(chunkID)
	if err != nil {
		return ectx.JSON(http.StatusBadRequest, "invalid uuid") // TODO: エラーメッセージ
	}

	return ectx.File(s.Filepath(id))
}

// WARNING: WriteChunkRequest.Data 以外をパースする
func parseRequst(ectx echo.Context) (*WriteChunkRequest, error) {
	chunkID := ectx.Param(PARAM_CHUNK_ID)
	id, err := uuid.FromString(chunkID)
	if err != nil {
		return nil, ectx.JSON(http.StatusBadRequest, "invalid uuid") // TODO: エラーメッセージ
	}

	prec := ectx.QueryParam(WRITE_QUERY_PRE_CHECKSUM)
	preChecksum, err := strconv.ParseUint(prec, 16, 64)
	if err != nil {
		return nil, ectx.JSON(http.StatusBadRequest, "invalid "+WRITE_QUERY_PRE_CHECKSUM) // TODO: エラーメッセージ
	}

	postc := ectx.QueryParam(WRITE_QUERY_POST_CHECKSUM)
	postChecksum, err := strconv.ParseUint(postc, 16, 64)
	if err != nil {
		return nil, ectx.JSON(http.StatusBadRequest, "invalid "+WRITE_QUERY_POST_CHECKSUM) // TODO: エラーメッセージ
	}

	o := ectx.QueryParam(WRITE_QUERY_OFFSET)
	offset, err := strconv.ParseInt(o, 16, 64)
	if err != nil {
		return nil, ectx.JSON(http.StatusBadRequest, "invalid "+WRITE_QUERY_OFFSET) // TODO: エラーメッセージ
	}

	nodes := strings.Split(ectx.QueryParam(WRITE_QUERY_REPLICA_NODES), ",")

	req := &WriteChunkRequest{
		ID:           id,
		Offset:       offset,
		PreChecksum:  preChecksum,
		PostChecksum: postChecksum,
		Nodes:        nodes,
	}

	return req, nil
}

func (s ChunkStoreService) WriteChunk(ectx echo.Context) error {
	req, err := parseRequst(ectx)
	if err != nil {
		return err
	}

	dataFile, err := ectx.FormFile(WRITE_FORM_DATA)
	if err != nil {
		return ectx.JSON(http.StatusBadRequest, "invalid "+WRITE_FORM_DATA) // TODO: エラーメッセージ
	}

	if s.MaxChunkSize-req.Offset < dataFile.Size {
		return ectx.JSON(http.StatusRequestEntityTooLarge, nil)
	}

	data, err := dataFile.Open()
	if err != nil {
		log.Printf("Open(): %s", err.Error())
		return ectx.JSON(http.StatusBadRequest, nil) // TODO: エラーメッセージ
	}
	defer data.Close()

	remote := OpenRemoteChunkNode()
	node := OpenChunkStoreNode(func(id uuid.UUID) (ChunkInterface, error) {
		return OpenFileChunk(s.Filepath(id), s.MaxChunkSize)
	}, remote)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := node.WriteChunk(ctx, req)

	return ectx.JSON(http.StatusCreated, res)
}
