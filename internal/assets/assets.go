// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package assets

import (
	_ "embed"
)

var (
	//go:embed api-schemas/cloud.json
	CloudOpenapiSchema []byte

	//go:embed api-schemas/cloud_v2.json
	CloudV2OpenapiSchema []byte

	//go:embed api-schemas/me.json
	MeOpenapiSchema []byte

	//go:embed api-schemas/baremetal.json
	BaremetalOpenapiSchema []byte

	//go:embed api-schemas/dedicatedceph.json
	DedicatedcephOpenapiSchema []byte

	//go:embed api-schemas/dedicatednasha.json
	DedicatednashaOpenapiSchema []byte

	//go:embed api-schemas/domain.json
	DomainOpenapiSchema []byte

	//go:embed api-schemas/emailmxplan.json
	EmailmxplanOpenapiSchema []byte

	//go:embed api-schemas/emailpro.json
	EmailproOpenapiSchema []byte

	//go:embed api-schemas/hostingprivatedatabase.json
	HostingprivatedatabaseOpenapiSchema []byte

	//go:embed api-schemas/iam.json
	IamOpenapiSchema []byte

	//go:embed api-schemas/ip.json
	IpOpenapiSchema []byte

	//go:embed api-schemas/iploadbalancing.json
	IploadbalancingOpenapiSchema []byte

	//go:embed api-schemas/ldp.json
	LdpOpenapiSchema []byte

	//go:embed api-schemas/overthebox.json
	OvertheboxOpenapiSchema []byte

	//go:embed api-schemas/ovhcloudconnect.json
	OvhcloudconnectOpenapiSchema []byte

	//go:embed api-schemas/packxdsl.json
	PackxdslOpenapiSchema []byte

	//go:embed api-schemas/sms.json
	SmsOpenapiSchema []byte

	//go:embed api-schemas/sslgateway.json
	SslgatewayOpenapiSchema []byte

	//go:embed api-schemas/storagenetapp.json
	StoragenetappOpenapiSchema []byte

	//go:embed api-schemas/telephony.json
	TelephonyOpenapiSchema []byte

	//go:embed api-schemas/vmwareclouddirectorbackup.json
	VmwareclouddirectorbackupOpenapiSchema []byte

	//go:embed api-schemas/vmwareclouddirectororganization.json
	VmwareclouddirectororganizationOpenapiSchema []byte

	//go:embed api-schemas/vps.json
	VpsOpenapiSchema []byte

	//go:embed api-schemas/vrack.json
	VrackOpenapiSchema []byte

	//go:embed api-schemas/vrackservices.json
	VrackservicesOpenapiSchema []byte

	//go:embed api-schemas/webhosting.json
	WebhostingOpenapiSchema []byte

	//go:embed api-schemas/xdsl.json
	XdslOpenapiSchema []byte
)
