package config

import (
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

func SetupMidtrans() snap.Client {
	var s snap.Client
	s.New("SB-Mid-server-0M3ckIvV22lFcF1Gl25R-LXW", midtrans.Sandbox) // Ganti dengan Server Key production jika sudah live
	return s
}
