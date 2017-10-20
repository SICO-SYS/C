/*

LICENSE:  MIT
Author:   sine
Email:    sinerwr@gmail.com

*/

package controller

import (
	"encoding/json"
	"github.com/getsentry/raven-go"
	"golang.org/x/net/context"

	"github.com/SiCo-Ops/Pb"
	"github.com/SiCo-Ops/dao/mongo"
	"github.com/SiCo-Ops/public"
)

type HookService struct{}

type hook struct {
	ID         string
	Name       string
	Belong     string
	Type       string
	CreateTime string
}

type HookResponse struct {
	Project  string `json:"project"`
	Branch   string `json:"branch"`
	CommitID string `json:"commitid"`
	Tag      string `json:"tag"`
	Time     string `json:"time"`
}

type ReceiveHookGithub struct {
}

type ReceiveHookTravis struct{}

type ReceiveHookDockerhub struct {
	PushData    ReceiveHookDockerhubPushData   `json:"push_data" bson:"push_data"`
	CallbackURL string                         `json:"callback_url" bson:"callback_url"`
	Repository  ReceiveHookDockerhubRepository `json:"repository" bson:"repository"`
}

type ReceiveHookDockerhubPushData struct {
	PushAt int64    `json:"pushed_at" bson:"pushed_at"`
	Images []string `json:"images" bson:"images"`
	Tag    string   `json:"tag" bson:"tag"`
	Pusher string   `json:"pusher" bson:"pusher"`
}

type ReceiveHookDockerhubRepository struct {
	Status          string `json:"status" bson:"status"`
	Description     string `json:"description" bson:"description"`
	IsTrusted       bool   `json:"is_trusted" bson:"is_trusted"`
	FullDescription string `json:"full_description" bson:"full_description"`
	RepoURL         string `json:"repo_url" bson:"repo_url"`
	Owner           string `json:"owner" bson:"owner"`
	IsOfficial      bool   `json:"is_official" bson:"is_official"`
	IsPrivate       bool   `json:"is_private" bson:"is_private"`
	Name            string `json:"name" bson:"name"`
	Namespace       string `json:"namespace" bson:"namespace"`
	StarCount       int64  `json:"star_count" bson:"star_count"`
	CommentCount    int64  `json:"comment_count" bson:"comment_count"`
	DateCreated     int64  `json:"date_created" bson:"date_created"`
	Dockerfile      string `json:"dockerfile" bson:"dockerfile"`
	RepoName        string `json:"repo_name" bson:"repo_name"`
}

func (h *HookService) AuthRPC(ctx context.Context, in *pb.HookAuthCall) (*pb.HookAuthBack, error) {
	hookName := in.Hookname
	hookres, hookerr := mongo.FindOne(hookDB, mongo.CollectionHookName(), map[string]string{"name": hookName, "belong": in.Id})
	if hookerr != nil {
		return &pb.HookAuthBack{Code: 205}, nil
	}
	if hookres == nil {
		return &pb.HookAuthBack{Code: 0}, nil
	}
	hookid, _ := hookres["id"].(string)
	return &pb.HookAuthBack{Code: 0, Hookid: hookid}, nil
}

func (h *HookService) CreateRPC(ctx context.Context, in *pb.HookCreateCall) (*pb.HookCreateBack, error) {
	v := &hook{}
	v.Belong = in.Id
	for i := 0; true; i++ {
		v.ID = public.GenerateHexString()
		v.Name = public.GenerateHexString()
		v.Type = in.Hooktype
		v.CreateTime = public.CurrentUTCFormat()
		err := mongo.Insert(hookDB, mongo.CollectionHookName(), v)
		if err != nil {
			if i <= 4 {
				continue
			}
			raven.CaptureError(err, nil)
			return &pb.HookCreateBack{Code: 305}, nil
		}
		break
	}
	return &pb.HookCreateBack{Code: 0, Hookname: v.Name}, nil
}

func (h *HookService) QueryRPC(ctx context.Context, in *pb.HookQueryCall) (*pb.HookQueryBack, error) {
	hookName := in.Hookname
	r, err := mongo.FindOne(hookDB, mongo.CollectionHookName(), map[string]string{"name": hookName})
	if err != nil {
		return &pb.HookQueryBack{Code: 205}, nil
	}
	if r != nil {
		hookid, _ := r["id"].(string)
		hookType, _ := r["type"].(string)
		belong, _ := r["belong"].(string)
		return &pb.HookQueryBack{Code: 0, Belong: belong, Hookid: hookid, Hooktype: hookType}, nil
	}
	return &pb.HookQueryBack{Code: 0}, nil
}

func (h *HookService) UpdateNameRPC(ctx context.Context, in *pb.HookUpdateNameCall) (*pb.HookUpdateNameBack, error) {
	return &pb.HookUpdateNameBack{Code: 0}, nil
}

func (h *HookService) ReceiveRPC(ctx context.Context, in *pb.HookReceiveCall) (*pb.HookReceiveBack, error) {
	hookType := in.Hooktype
	payload := in.Payload
	switch hookType {
	case "dockerhub":
		v := &ReceiveHookDockerhub{}
		err := json.Unmarshal(payload, v)
		if err != nil {
			return &pb.HookReceiveBack{Code: 5000}, nil
		}
		err = mongo.Insert(hookDB, mongo.CollectionHookReceiveName(hookType), v)
		if err != nil {
			return &pb.HookReceiveBack{Code: 205}, nil
		}
		response := &HookResponse{}
		response.Project = v.Repository.Name
		response.Tag = v.PushData.Tag
		param, _ := json.Marshal(response)
		return &pb.HookReceiveBack{Code: 0, Params: param}, nil
	default:
		return &pb.HookReceiveBack{Code: 0, Params: nil}, nil
	}
	return &pb.HookReceiveBack{Code: 0, Params: nil}, nil
}
