package service

import (
	"blog_service/global"
	"blog_service/internal/dao"
	"context"
)

type Service struct {
	ctx context.Context
	dao *dao.Dao
}

func New(ctx context.Context) Service{
	svc := Service{ctx:ctx}
	svc.dao = dao.New(global.DBEngine)
}
