package entity

var (
	ProviderToID = map[Provider]int{
		ProviderGmail:   1,
		ProviderOutlook: 2,
		ProviderYandex:  3,
		ProviderMailru:  4,
		ProviderCustom:  5,
	}
	IDToProvider = map[int]Provider{
		1: ProviderGmail,
		2: ProviderOutlook,
		3: ProviderYandex,
		4: ProviderMailru,
		5: ProviderCustom,
	}

	ProtocolToID = map[Protocol]int{
		ProtocolIMAP: 1,
		ProtocolPOP3: 2,
	}
	IDToProtocol = map[int]Protocol{
		1: ProtocolIMAP,
		2: ProtocolPOP3,
	}

	AuthTypeToID = map[AuthType]int{
		AuthTypePlain:  1,
		AuthTypeOAuth2: 2,
	}
	IDToAuthType = map[int]AuthType{
		1: AuthTypePlain,
		2: AuthTypeOAuth2,
	}

	SyncStatusToID = map[SyncStatus]int{
		SyncStatusPending: 1,
		SyncStatusRunning: 2,
		SyncStatusSuccess: 3,
		SyncStatusFailed:  4,
	}
	IDToSyncStatus = map[int]SyncStatus{
		1: SyncStatusPending,
		2: SyncStatusRunning,
		3: SyncStatusSuccess,
		4: SyncStatusFailed,
	}
)
