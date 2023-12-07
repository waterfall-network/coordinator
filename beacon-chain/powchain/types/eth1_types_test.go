package types

import (
	"math/big"
	"reflect"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gethTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
)

func Test_headerToHeaderInfo(t *testing.T) {
	nr0 := uint64(0)
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
				Number: &nr0,
				Time:   2345,
			}},
			want: &HeaderInfo{
				Number: big.NewInt(0),
				Hash:   common.Hash{0xBA, 0xB6, 0x58, 0x2B, 0x44, 0xBF, 0x34, 0x7C, 0x75, 0xAA, 0x0B, 0xD2, 0xC3, 0x00, 0x06, 0xA2, 0xD5, 0x10, 0xB7, 0xAF, 0x85, 0x4D, 0x4F, 0xB1, 0x8B, 0xE0, 0x22, 0x24, 0xC0, 0xB8, 0xC3, 0xB8},
				Time:   2345,
			},
		},
		{
			name: "nil number",
			args: args{hdr: &gethTypes.Header{
				Number: &nr0,
				Height: 5,
				Time:   2345,
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
