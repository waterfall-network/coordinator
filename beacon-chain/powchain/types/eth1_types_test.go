package types

import (
	"math/big"
	"reflect"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gethTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
)

func Test_headerToHeaderInfo(t *testing.T) {
	nr_0 := uint64(500)
	type args struct {
		hdr *gethTypes.Header
	}
	tests := []struct {
		name    string
		args    args
		want    *HeaderInfo
		wantErr bool
	}{
		{
			name: "OK",
			args: args{hdr: &gethTypes.Header{
				Number: &nr_0,
				Time:   2345,
			}},
			want: &HeaderInfo{
				Number: big.NewInt(500),
				Hash:   common.Hash{0xCD, 0x53, 0x3C, 0xB5, 0xB9, 0x65, 0xD9, 0xFE, 0xF1, 0xAC, 0xDB, 0x44, 0x5D, 0x6E, 0x70, 0x7F, 0xC8, 0x2B, 0xE2, 0x5F, 0x34, 0xB1, 0x94, 0x9A, 0xBD, 0x87, 0x34, 0x79, 0x6E, 0x4E, 0x18, 0x42},
				Time:   2345,
			},
		},
		{
			name: "nil number",
			args: args{hdr: &gethTypes.Header{
				Time: 2345,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HeaderToHeaderInfo(tt.args.hdr)
			if (err != nil) != tt.wantErr {
				t.Errorf("headerToHeaderInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headerToHeaderInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
