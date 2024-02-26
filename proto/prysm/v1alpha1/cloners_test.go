package eth_test

import (
	"math/rand"
	"reflect"
	"testing"

	v1alpha1 "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

func TestCopyETH1Data(t *testing.T) {
	data := genEth1Data()

	got := v1alpha1.CopyETH1Data(data)
	if !reflect.DeepEqual(got, data) {
		t.Errorf("CopyETH1Data() = %v, want %v", got, data)
	}
}

func TestCopySpineData(t *testing.T) {
	data := genSpineData()

	got := v1alpha1.CopySpineData(data)
	if !reflect.DeepEqual(got, data) {
		t.Errorf("CopyETH1Data() = %v, want %v", got, data)
	}
}

func TestCopyPendingAttestation(t *testing.T) {
	pa := genPendingAttestation()

	got := v1alpha1.CopyPendingAttestation(pa)
	if !reflect.DeepEqual(got, pa) {
		t.Errorf("CopyPendingAttestation() = %v, want %v", got, pa)
	}
}

func TestCopyAttestation(t *testing.T) {
	att := genAttestation()

	got := v1alpha1.CopyAttestation(att)
	if !reflect.DeepEqual(got, att) {
		t.Errorf("CopyAttestation() = %v, want %v", got, att)
	}
}
func TestCopyAttestationData(t *testing.T) {
	att := genAttData()

	got := v1alpha1.CopyAttestationData(att)
	if !reflect.DeepEqual(got, att) {
		t.Errorf("CopyAttestationData() = %v, want %v", got, att)
	}
}

func TestCopyCheckpoint(t *testing.T) {
	cp := genCheckpoint()

	got := v1alpha1.CopyCheckpoint(cp)
	if !reflect.DeepEqual(got, cp) {
		t.Errorf("CopyCheckpoint() = %v, want %v", got, cp)
	}
}

func TestCopySignedBeaconBlock(t *testing.T) {
	blk := genSignedBeaconBlock()

	got := v1alpha1.CopySignedBeaconBlock(blk)
	if !reflect.DeepEqual(got, blk) {
		t.Errorf("CopySignedBeaconBlock() = %v, want %v", got, blk)
	}
}

func TestCopyBeaconBlock(t *testing.T) {
	blk := genBeaconBlock()

	got := v1alpha1.CopyBeaconBlock(blk)
	if !reflect.DeepEqual(got, blk) {
		t.Errorf("CopyBeaconBlock() = %v, want %v", got, blk)
	}
}

func TestCopyBeaconBlockBody(t *testing.T) {
	body := genBeaconBlockBody()

	got := v1alpha1.CopyBeaconBlockBody(body)
	if !reflect.DeepEqual(got, body) {
		t.Errorf("CopyBeaconBlockBody() = %v, want %v", got, body)
	}
}

func TestCopySignedBeaconBlockAltair(t *testing.T) {
	sbb := genSignedBeaconBlockAltair()

	got := v1alpha1.CopySignedBeaconBlockAltair(sbb)
	if !reflect.DeepEqual(got, sbb) {
		t.Errorf("CopySignedBeaconBlockAltair() = %v, want %v", got, sbb)
	}
}

func TestCopyBeaconBlockAltair(t *testing.T) {
	b := genBeaconBlockAltair()

	got := v1alpha1.CopyBeaconBlockAltair(b)
	if !reflect.DeepEqual(got, b) {
		t.Errorf("CopyBeaconBlockAltair() = %v, want %v", got, b)
	}
}

func TestCopyBeaconBlockBodyAltair(t *testing.T) {
	bb := genBeaconBlockBodyAltair()

	got := v1alpha1.CopyBeaconBlockBodyAltair(bb)
	if !reflect.DeepEqual(got, bb) {
		t.Errorf("CopyBeaconBlockBodyAltair() = %v, want %v", got, bb)
	}
}

func TestCopyProposerSlashings(t *testing.T) {
	ps := genProposerSlashings(10)

	got := v1alpha1.CopyProposerSlashings(ps)
	if !reflect.DeepEqual(got, ps) {
		t.Errorf("CopyProposerSlashings() = %v, want %v", got, ps)
	}
}

func TestCopyProposerSlashing(t *testing.T) {
	ps := genProposerSlashing()

	got := v1alpha1.CopyProposerSlashing(ps)
	if !reflect.DeepEqual(got, ps) {
		t.Errorf("CopyProposerSlashing() = %v, want %v", got, ps)
	}
}

func TestCopySignedBeaconBlockHeader(t *testing.T) {
	sbh := genSignedBeaconBlockHeader()

	got := v1alpha1.CopySignedBeaconBlockHeader(sbh)
	if !reflect.DeepEqual(got, sbh) {
		t.Errorf("CopySignedBeaconBlockHeader() = %v, want %v", got, sbh)
	}
}

func TestCopyBeaconBlockHeader(t *testing.T) {
	bh := genBeaconBlockHeader()

	got := v1alpha1.CopyBeaconBlockHeader(bh)
	if !reflect.DeepEqual(got, bh) {
		t.Errorf("CopyBeaconBlockHeader() = %v, want %v", got, bh)
	}
}

func TestCopyAttesterSlashings(t *testing.T) {
	as := genAttesterSlashings(10)

	got := v1alpha1.CopyAttesterSlashings(as)
	if !reflect.DeepEqual(got, as) {
		t.Errorf("CopyAttesterSlashings() = %v, want %v", got, as)
	}
}

func TestCopyIndexedAttestation(t *testing.T) {
	ia := genIndexedAttestation()

	got := v1alpha1.CopyIndexedAttestation(ia)
	if !reflect.DeepEqual(got, ia) {
		t.Errorf("CopyIndexedAttestation() = %v, want %v", got, ia)
	}
}

func TestCopyAttestations(t *testing.T) {
	atts := genAttestations(10)

	got := v1alpha1.CopyAttestations(atts)
	if !reflect.DeepEqual(got, atts) {
		t.Errorf("CopyAttestations() = %v, want %v", got, atts)
	}
}

func TestCopyDeposits(t *testing.T) {
	d := genDeposits(10)

	got := v1alpha1.CopyDeposits(d)
	if !reflect.DeepEqual(got, d) {
		t.Errorf("CopyDeposits() = %v, want %v", got, d)
	}
}

func TestCopyDeposit(t *testing.T) {
	d := genDeposit()

	got := v1alpha1.CopyDeposit(d)
	if !reflect.DeepEqual(got, d) {
		t.Errorf("CopyDeposit() = %v, want %v", got, d)
	}
}

func TestCopyDepositData(t *testing.T) {
	dd := genDepositData()

	got := v1alpha1.CopyDepositData(dd)
	if !reflect.DeepEqual(got, dd) {
		t.Errorf("CopyDepositData() = %v, want %v", got, dd)
	}
}

func TestCopyVoluntaryExits(t *testing.T) {
	sv := genVoluntaryExits(10)

	got := v1alpha1.CopyVoluntaryExits(sv)
	if !reflect.DeepEqual(got, sv) {
		t.Errorf("CopyVoluntaryExits() = %v, want %v", got, sv)
	}
}

func TestCopyVoluntaryExit(t *testing.T) {
	sv := genVoluntaryExit()

	got := v1alpha1.CopyVoluntaryExit(sv)
	if !reflect.DeepEqual(got, sv) {
		t.Errorf("CopyVoluntaryExit() = %v, want %v", got, sv)
	}
}

func TestCopyValidator(t *testing.T) {
	v := genValidator()

	got := v1alpha1.CopyValidator(v)
	if !reflect.DeepEqual(got, v) {
		t.Errorf("CopyValidator() = %v, want %v", got, v)
	}
}

func TestCopySyncCommitteeMessage(t *testing.T) {
	scm := genSyncCommitteeMessage()

	got := v1alpha1.CopySyncCommitteeMessage(scm)
	if !reflect.DeepEqual(got, scm) {
		t.Errorf("CopySyncCommitteeMessage() = %v, want %v", got, scm)
	}
}

func TestCopySyncCommitteeContribution(t *testing.T) {
	scc := genSyncCommitteeContribution()

	got := v1alpha1.CopySyncCommitteeContribution(scc)
	if !reflect.DeepEqual(got, scc) {
		t.Errorf("CopySyncCommitteeContribution() = %v, want %v", got, scc)
	}
}

func TestCopySyncAggregate(t *testing.T) {
	sa := genSyncAggregate()

	got := v1alpha1.CopySyncAggregate(sa)
	if !reflect.DeepEqual(got, sa) {
		t.Errorf("CopySyncAggregate() = %v, want %v", got, sa)
	}
}

func TestCopyPendingAttestationSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []*v1alpha1.PendingAttestation
	}{
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "empty",
			input: []*v1alpha1.PendingAttestation{},
		},
		{
			name: "correct copy",
			input: []*v1alpha1.PendingAttestation{
				genPendingAttestation(),
				genPendingAttestation(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v1alpha1.CopyPendingAttestationSlice(tt.input); !reflect.DeepEqual(got, tt.input) {
				t.Errorf("CopyPendingAttestationSlice() = %v, want %v", got, tt.input)
			}
		})
	}
}

func bytes() []byte {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 32; i++ {
		if b[i] == 0x00 {
			b[i] = uint8(rand.Int())
		}
	}
	return b
}

func genAttestation() *v1alpha1.Attestation {
	return &v1alpha1.Attestation{
		AggregationBits: bytes(),
		Data:            genAttData(),
		Signature:       bytes(),
	}
}

func genAttestations(num int) []*v1alpha1.Attestation {
	atts := make([]*v1alpha1.Attestation, num)
	for i := 0; i < num; i++ {
		atts[i] = genAttestation()
	}
	return atts
}

func genAttData() *v1alpha1.AttestationData {
	return &v1alpha1.AttestationData{
		Slot:            1,
		CommitteeIndex:  2,
		BeaconBlockRoot: bytes(),
		Source:          genCheckpoint(),
		Target:          genCheckpoint(),
	}
}

func genCheckpoint() *v1alpha1.Checkpoint {
	return &v1alpha1.Checkpoint{
		Epoch: 1,
		Root:  bytes(),
	}
}

func genEth1Data() *v1alpha1.Eth1Data {
	return &v1alpha1.Eth1Data{
		DepositRoot:  bytes(),
		DepositCount: 4,
		BlockHash:    bytes(),
		Candidates:   bytes(),
	}
}

func genSpineData() *v1alpha1.SpineData {
	spines := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		gwatCommon.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
	}
	prefix := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333"),
		gwatCommon.HexToHash("0x4444444444444444444444444444444444444444444444444444444444444444"),
	}
	finalization := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x5555555555555555555555555555555555555555555555555555555555555555"),
		gwatCommon.HexToHash("0x6666666666666666666666666666666666666666666666666666666666666666"),
	}
	cpFinalized := gwatCommon.HashArray{
		gwatCommon.HexToHash("0x5555555555555555555555555555555555555555555555555555555555555555"),
		gwatCommon.HexToHash("0x6666666666666666666666666666666666666666666666666666666666666666"),
	}

	parentSpines := []*v1alpha1.SpinesSeq{
		{Spines: spines.ToBytes()},
		{Spines: gwatCommon.HashArray{
			gwatCommon.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
			gwatCommon.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
			gwatCommon.HexToHash("0x7777777777777777777777777777777777777777777777777777777777777777"),
			gwatCommon.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffff0101010101010101"),
		}.ToBytes(),
		},
	}

	return &v1alpha1.SpineData{
		Spines:       spines.ToBytes(),
		Prefix:       prefix.ToBytes(),
		Finalization: finalization.ToBytes(),
		CpFinalized:  cpFinalized.ToBytes(),
		ParentSpines: parentSpines,
	}
}

func genPendingAttestation() *v1alpha1.PendingAttestation {
	return &v1alpha1.PendingAttestation{
		AggregationBits: bytes(),
		Data:            genAttData(),
		InclusionDelay:  3,
		ProposerIndex:   5,
	}
}

func genSignedBeaconBlock() *v1alpha1.SignedBeaconBlock {
	return &v1alpha1.SignedBeaconBlock{
		Block:     genBeaconBlock(),
		Signature: bytes(),
	}
}

func genBeaconBlock() *v1alpha1.BeaconBlock {
	return &v1alpha1.BeaconBlock{
		Slot:          4,
		ProposerIndex: 5,
		ParentRoot:    bytes(),
		StateRoot:     bytes(),
		Body:          genBeaconBlockBody(),
	}
}

func genBeaconBlockBody() *v1alpha1.BeaconBlockBody {
	return &v1alpha1.BeaconBlockBody{
		RandaoReveal:      bytes(),
		Eth1Data:          genEth1Data(),
		Graffiti:          bytes(),
		ProposerSlashings: genProposerSlashings(5),
		AttesterSlashings: genAttesterSlashings(5),
		Attestations:      genAttestations(5),
		Deposits:          genDeposits(5),
		VoluntaryExits:    genVoluntaryExits(5),
		Withdrawals:       genWithdrawals(5),
	}
}

func genWithdrawals(num int) []*v1alpha1.Withdrawal {
	d := make([]*v1alpha1.Withdrawal, num)
	for i := 0; i < num; i++ {
		d[i] = genWithdrawal()
	}

	return d
}

func genWithdrawal() *v1alpha1.Withdrawal {
	return &v1alpha1.Withdrawal{
		PublicKey:      bytes(),
		ValidatorIndex: 12345,
		Amount:         3200,
		InitTxHash:     bytes(),
		Epoch:          99999,
	}
}

func genProposerSlashing() *v1alpha1.ProposerSlashing {
	return &v1alpha1.ProposerSlashing{
		Header_1: genSignedBeaconBlockHeader(),
		Header_2: genSignedBeaconBlockHeader(),
	}
}

func genProposerSlashings(num int) []*v1alpha1.ProposerSlashing {
	ps := make([]*v1alpha1.ProposerSlashing, num)
	for i := 0; i < num; i++ {
		ps[i] = genProposerSlashing()
	}
	return ps
}

func genAttesterSlashing() *v1alpha1.AttesterSlashing {
	return &v1alpha1.AttesterSlashing{
		Attestation_1: genIndexedAttestation(),
		Attestation_2: genIndexedAttestation(),
	}
}

func genIndexedAttestation() *v1alpha1.IndexedAttestation {
	return &v1alpha1.IndexedAttestation{
		AttestingIndices: []uint64{1, 2, 3},
		Data:             genAttData(),
		Signature:        bytes(),
	}
}

func genAttesterSlashings(num int) []*v1alpha1.AttesterSlashing {
	as := make([]*v1alpha1.AttesterSlashing, num)
	for i := 0; i < num; i++ {
		as[i] = genAttesterSlashing()
	}
	return as
}

func genBeaconBlockHeader() *v1alpha1.BeaconBlockHeader {
	return &v1alpha1.BeaconBlockHeader{
		Slot:          10,
		ProposerIndex: 15,
		ParentRoot:    bytes(),
		StateRoot:     bytes(),
		BodyRoot:      bytes(),
	}
}

func genSignedBeaconBlockHeader() *v1alpha1.SignedBeaconBlockHeader {
	return &v1alpha1.SignedBeaconBlockHeader{
		Header:    genBeaconBlockHeader(),
		Signature: bytes(),
	}
}

func genDepositData() *v1alpha1.Deposit_Data {
	return &v1alpha1.Deposit_Data{
		PublicKey:             bytes(),
		CreatorAddress:        bytes(),
		WithdrawalCredentials: bytes(),
		Amount:                20000,
		Signature:             bytes(),
		InitTxHash:            bytes(),
	}
}

func genDeposit() *v1alpha1.Deposit {
	return &v1alpha1.Deposit{
		Data:  genDepositData(),
		Proof: [][]byte{bytes(), bytes(), bytes(), bytes()},
	}
}

func genDeposits(num int) []*v1alpha1.Deposit {
	d := make([]*v1alpha1.Deposit, num)
	for i := 0; i < num; i++ {
		d[i] = genDeposit()
	}
	return d
}

func genVoluntaryExit() *v1alpha1.VoluntaryExit {
	return &v1alpha1.VoluntaryExit{
		Epoch:          5432,
		ValidatorIndex: 888888,
		InitTxHash:     bytes(),
	}
}

func genVoluntaryExits(num int) []*v1alpha1.VoluntaryExit {
	sv := make([]*v1alpha1.VoluntaryExit, num)
	for i := 0; i < num; i++ {
		sv[i] = genVoluntaryExit()
	}
	return sv
}

func genValidator() *v1alpha1.Validator {
	return &v1alpha1.Validator{
		PublicKey:                  bytes(),
		WithdrawalCredentials:      bytes(),
		EffectiveBalance:           12345,
		Slashed:                    true,
		ActivationEligibilityEpoch: 14322,
		ActivationEpoch:            14325,
		ExitEpoch:                  23425,
		WithdrawableEpoch:          30000,
		ActivationHash:             bytes(),
		ExitHash:                   bytes(),
		WithdrawalOps: []*v1alpha1.WithdrawalOp{{
			Amount: 3200,
			Hash:   bytes(),
			Slot:   55,
		}},
		CreatorAddress: bytes(),
	}
}

func genSyncCommitteeContribution() *v1alpha1.SyncCommitteeContribution {
	return &v1alpha1.SyncCommitteeContribution{
		Slot:              12333,
		BlockRoot:         bytes(),
		SubcommitteeIndex: 4444,
		AggregationBits:   bytes(),
		Signature:         bytes(),
	}
}

func genSyncAggregate() *v1alpha1.SyncAggregate {
	return &v1alpha1.SyncAggregate{
		SyncCommitteeBits:      bytes(),
		SyncCommitteeSignature: bytes(),
	}
}

func genBeaconBlockBodyAltair() *v1alpha1.BeaconBlockBodyAltair {
	return &v1alpha1.BeaconBlockBodyAltair{
		RandaoReveal:      bytes(),
		Eth1Data:          genEth1Data(),
		Graffiti:          bytes(),
		ProposerSlashings: genProposerSlashings(5),
		AttesterSlashings: genAttesterSlashings(5),
		Attestations:      genAttestations(10),
		Deposits:          genDeposits(5),
		VoluntaryExits:    genVoluntaryExits(12),
		SyncAggregate:     genSyncAggregate(),
		Withdrawals:       genWithdrawals(5),
	}
}

func genBeaconBlockAltair() *v1alpha1.BeaconBlockAltair {
	return &v1alpha1.BeaconBlockAltair{
		Slot:          123455,
		ProposerIndex: 55433,
		ParentRoot:    bytes(),
		StateRoot:     bytes(),
		Body:          genBeaconBlockBodyAltair(),
	}
}

func genSignedBeaconBlockAltair() *v1alpha1.SignedBeaconBlockAltair {
	return &v1alpha1.SignedBeaconBlockAltair{
		Block:     genBeaconBlockAltair(),
		Signature: bytes(),
	}
}

func genSyncCommitteeMessage() *v1alpha1.SyncCommitteeMessage {
	return &v1alpha1.SyncCommitteeMessage{
		Slot:           424555,
		BlockRoot:      bytes(),
		ValidatorIndex: 5443,
		Signature:      bytes(),
	}
}
