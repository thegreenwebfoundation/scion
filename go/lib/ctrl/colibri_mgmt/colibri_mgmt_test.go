// Copyright 2020 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package colibri_mgmt_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/scionproto/scion/go/lib/ctrl/colibri_mgmt"
	"github.com/scionproto/scion/go/lib/xtest"
	"github.com/scionproto/scion/go/proto"
)

func TestSerializeRoot(t *testing.T) {
	root := &colibri_mgmt.ColibriRequestPayload{
		Timestamp: 42,
		Which:     proto.ColibriRequestPayload_Which_unset,
	}
	buffer, err := root.PackRoot()
	require.NoError(t, err)
	require.Len(t, buffer, 9)
	otherRoot, err := colibri_mgmt.NewFromRaw(buffer)
	require.NoError(t, err)
	require.Equal(t, root.Which, otherRoot.Which)
	otherBuffer, err := otherRoot.PackRoot()
	require.NoError(t, err)
	require.Equal(t, buffer, otherBuffer)
}

// tests serialization for all types of requests
func TestSerializeRequest(t *testing.T) {
	newSegmentSetup := func() *colibri_mgmt.SegmentSetup {
		return &colibri_mgmt.SegmentSetup{
			Base:     newSegmentBase(),
			MinBW:    1,
			MaxBW:    2,
			SplitCls: 3,
			StartProps: colibri_mgmt.PathEndProps{
				Local:    true,
				Transfer: false,
			},
			EndProps: colibri_mgmt.PathEndProps{
				Local:    false,
				Transfer: true,
			},
			InfoField: xtest.MustParseHexString("0123456789abcdef"),
			AllocationTrail: []*colibri_mgmt.AllocationBead{
				{
					AllocBW: 5,
					MaxBW:   6,
				},
			},
		}
	}

	testCases := map[string]struct {
		Request *colibri_mgmt.Request
	}{
		"setup": {
			Request: &colibri_mgmt.Request{
				Which:        proto.Request_Which_segmentSetup,
				SegmentSetup: newSegmentSetup(),
			},
		},
		"renewal": {
			Request: &colibri_mgmt.Request{
				Which:          proto.Request_Which_segmentRenewal,
				SegmentRenewal: newSegmentSetup(),
			},
		},
		"teles_setup": {
			Request: &colibri_mgmt.Request{
				Which: proto.Request_Which_segmentTelesSetup,
				SegmentTelesSetup: &colibri_mgmt.SegmentTelesSetup{
					Setup: newSegmentSetup(),
					BaseID: &colibri_mgmt.SegmentReservationID{
						ASID:   xtest.MustParseHexString("ff00cafe0001"),
						Suffix: xtest.MustParseHexString("deadbeef"),
					},
				},
			},
		},
		"teles_renewal": {
			Request: &colibri_mgmt.Request{
				Which: proto.Request_Which_segmentTelesRenewal,
				SegmentTelesRenewal: &colibri_mgmt.SegmentTelesSetup{
					Setup: newSegmentSetup(),
					BaseID: &colibri_mgmt.SegmentReservationID{
						ASID:   xtest.MustParseHexString("ff00cafe0001"),
						Suffix: xtest.MustParseHexString("deadbeef"),
					},
				},
			},
		},
		"teardown": {
			Request: &colibri_mgmt.Request{
				Which: proto.Request_Which_segmentTeardown,
				SegmentTeardown: &colibri_mgmt.SegmentTeardownReq{
					Base: newSegmentBase(),
				},
			},
		},
		"index_confirmation": {
			Request: &colibri_mgmt.Request{
				Which: proto.Request_Which_segmentIndexConfirmation,
				SegmentIndexConfirmation: &colibri_mgmt.SegmentIndexConfirmation{
					Base:  newSegmentBase(),
					State: proto.ReservationIndexState_active,
				},
			},
		},
		"cleanup": {
			Request: &colibri_mgmt.Request{
				Which: proto.Request_Which_segmentCleanup,
				SegmentCleanup: &colibri_mgmt.SegmentCleanup{
					Base: newSegmentBase(),
				},
			},
		},
		"e2esetup": {
			Request: &colibri_mgmt.Request{
				Which: proto.Request_Which_e2eSetup,
				E2ESetup: &colibri_mgmt.E2ESetup{
					Base: newE2EBase(),
					SegmentRsvs: []colibri_mgmt.SegmentReservationID{
						{
							ASID:   xtest.MustParseHexString("ff00cafe0001"),
							Suffix: xtest.MustParseHexString("deadbeef"),
						},
					},
					SegmentRsvASCount: []uint8{3},
					RequestedBW:       5,
					AllocationTrail:   []uint8{5},
					Which:             proto.E2ESetupReqData_Which_success,
					Success: &colibri_mgmt.E2ESetupReqSuccess{
						Token: xtest.MustParseHexString("0000"),
					},
				},
			},
		},
		"e2erenewal": {
			Request: &colibri_mgmt.Request{
				Which: proto.Request_Which_e2eRenewal,
				E2ERenewal: &colibri_mgmt.E2ESetup{
					Base: newE2EBase(),
					SegmentRsvs: []colibri_mgmt.SegmentReservationID{
						{
							ASID:   xtest.MustParseHexString("ff00cafe0001"),
							Suffix: xtest.MustParseHexString("00000001"),
						},
						{
							ASID:   xtest.MustParseHexString("ff00cafe0002"),
							Suffix: xtest.MustParseHexString("00000001"),
						},
					},
					SegmentRsvASCount: []uint8{3, 2},
					RequestedBW:       5,
					AllocationTrail:   []uint8{5, 4},
					Which:             proto.E2ESetupReqData_Which_failure,
					Failure: &colibri_mgmt.E2ESetupReqFailure{
						ErrorCode: 66,
					},
				},
			},
		},
		"e2ecleanup": {
			Request: &colibri_mgmt.Request{
				Which: proto.Request_Which_e2eCleanup,
				E2ECleanup: &colibri_mgmt.E2ECleanup{
					Base: newE2EBase(),
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := &colibri_mgmt.ColibriRequestPayload{
				Which:   proto.ColibriRequestPayload_Which_request,
				Request: tc.Request,
			}
			buffer, err := root.PackRoot()
			require.NoError(t, err)
			otherRoot, err := colibri_mgmt.NewFromRaw(buffer)
			require.NoError(t, err)
			otherBuffer, err := otherRoot.PackRoot()
			require.NoError(t, err)
			require.Equal(t, buffer, otherBuffer)
		})
	}
}

func TestSerializeResponse(t *testing.T) {
	newSegmentSetupResp := func() *colibri_mgmt.SegmentSetupRes {
		return &colibri_mgmt.SegmentSetupRes{
			Base:  newSegmentBase(),
			Which: proto.SegmentSetupResData_Which_token,
			Token: xtest.MustParseHexString("0000"),
		}
	}
	newE2ESetup := func() *colibri_mgmt.E2ESetupRes {
		return &colibri_mgmt.E2ESetupRes{
			Base:  newE2EBase(),
			Which: proto.E2ESetupResData_Which_failure,
			Failure: &colibri_mgmt.E2ESetupFailure{
				ErrorCode:       1,
				AllocationTrail: []uint8{1, 1, 2, 2},
			},
		}
	}
	testCases := map[string]struct {
		Response *colibri_mgmt.Response
	}{
		"setup failed": {
			Response: &colibri_mgmt.Response{
				Which: proto.Response_Which_segmentSetup,
				SegmentSetup: &colibri_mgmt.SegmentSetupRes{
					Which: proto.SegmentSetupResData_Which_failure,
					Failure: &colibri_mgmt.SegmentSetup{
						Base:     newSegmentBase(),
						MinBW:    1,
						MaxBW:    2,
						SplitCls: 3,
						StartProps: colibri_mgmt.PathEndProps{
							Local:    true,
							Transfer: false,
						},
						EndProps: colibri_mgmt.PathEndProps{
							Local:    false,
							Transfer: true,
						},
						AllocationTrail: []*colibri_mgmt.AllocationBead{
							{
								AllocBW: 5,
								MaxBW:   6,
							},
						},
					},
				},
				Accepted:  false,
				FailedHop: 3,
			},
		},
		"setup success": {
			Response: &colibri_mgmt.Response{
				Which:        proto.Response_Which_segmentSetup,
				SegmentSetup: newSegmentSetupResp(),
				Accepted:     true,
			},
		},
		"renewal": {
			Response: &colibri_mgmt.Response{
				Which:          proto.Response_Which_segmentRenewal,
				SegmentRenewal: newSegmentSetupResp(),
				Accepted:       true,
			},
		},
		"teardown": {
			Response: &colibri_mgmt.Response{
				Which: proto.Response_Which_segmentTeardown,
				SegmentTeardown: &colibri_mgmt.SegmentTeardownRes{
					Base:      newSegmentBase(),
					ErrorCode: 123,
				},
				Accepted:  false,
				FailedHop: 2,
			},
		},
		"index confirmation": {
			Response: &colibri_mgmt.Response{
				Which: proto.Response_Which_segmentIndexConfirmation,
				SegmentIndexConfirmation: &colibri_mgmt.SegmentIndexConfirmationRes{
					Base: newSegmentBase(),
				},
				Accepted: true,
			},
		},
		"segment cleanup": {
			Response: &colibri_mgmt.Response{
				Which: proto.Response_Which_segmentCleanup,
				SegmentCleanup: &colibri_mgmt.SegmentCleanupRes{
					Base:      newSegmentBase(),
					ErrorCode: 12,
				},
				Accepted: false,
			},
		},
		"e2e setup success": {
			Response: &colibri_mgmt.Response{
				Which: proto.Response_Which_e2eSetup,
				E2ESetup: &colibri_mgmt.E2ESetupRes{
					Base:  newE2EBase(),
					Which: proto.E2ESetupResData_Which_success,
					Success: &colibri_mgmt.E2ESetupSuccess{
						Token: xtest.MustParseHexString("0000"),
					},
				},
			},
		},
		"e2e setup failure": {
			Response: &colibri_mgmt.Response{
				Which: proto.Response_Which_e2eSetup,
				E2ESetup: &colibri_mgmt.E2ESetupRes{
					Base:  newE2EBase(),
					Which: proto.E2ESetupResData_Which_failure,
					Failure: &colibri_mgmt.E2ESetupFailure{
						ErrorCode:       1,
						AllocationTrail: []uint8{1, 1, 2, 2},
					},
				},
			},
		},
		"e2e renewal": {
			Response: &colibri_mgmt.Response{
				Which:      proto.Response_Which_e2eRenewal,
				E2ERenewal: newE2ESetup(),
				Accepted:   true,
			},
		},
		"e2e cleanup": {
			Response: &colibri_mgmt.Response{
				Which: proto.Response_Which_e2eCleanup,
				E2ECleanup: &colibri_mgmt.E2ECleanupRes{
					Base:      newE2EBase(),
					ErrorCode: 42,
				},
				Accepted: true,
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := &colibri_mgmt.ColibriRequestPayload{
				Which:    proto.ColibriRequestPayload_Which_response,
				Response: tc.Response,
			}
			buffer, err := root.PackRoot()
			require.NoError(t, err)
			otherRoot, err := colibri_mgmt.NewFromRaw(buffer)
			require.NoError(t, err)
			otherBuffer, err := otherRoot.PackRoot()
			require.NoError(t, err)
			require.Equal(t, buffer, otherBuffer)
		})
	}
}

func newSegmentBase() *colibri_mgmt.SegmentBase {
	return &colibri_mgmt.SegmentBase{
		ID: &colibri_mgmt.SegmentReservationID{
			ASID:   xtest.MustParseHexString("ff00cafe0001"),
			Suffix: xtest.MustParseHexString("deadbeef"),
		},
		Index: 10,
	}
}

func newE2EBase() *colibri_mgmt.E2EBase {
	return &colibri_mgmt.E2EBase{
		ID: &colibri_mgmt.E2EReservationID{
			ASID:   xtest.MustParseHexString("ff00cafe0001"),
			Suffix: xtest.MustParseHexString("0123456789abcdef0123"),
		},
		Index: 11,
	}
}
