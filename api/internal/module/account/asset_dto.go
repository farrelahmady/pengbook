package account

// AssetAccount is a posting asset account with its cached balance.
// Balance comes from account_balances (COALESCE to 0 when the row is
// missing, e.g. header accounts never have a balance row).
type AssetAccount struct {
	ID        int64   `json:"id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Balance   float64 `json:"balance"`
	IsPosting bool    `json:"isPosting"`
}

// AssetGroup groups asset posting accounts under their level-2 ancestor
// (e.g. 1.01.01.00 Bank, 1.01.02.00 Cash), mirroring the level-2 seeder.
type AssetGroup struct {
	ID           int64          `json:"id"`
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	Accounts     []AssetAccount `json:"accounts"`
	TotalBalance float64        `json:"totalBalance"`
}

// AssetSummary is the lightweight response for the Asset topbar.
// It carries only aggregates; the per-group detail lives in AssetGroups.
type AssetSummary struct {
	TotalAsset   float64 `json:"totalAsset"`
	CurrentAsset float64 `json:"currentAsset"`
	AccountCount int     `json:"accountCount"`
}

// AssetGroups is the detail response for the Asset list.
type AssetGroups struct {
	Groups []AssetGroup `json:"groups"`
}
