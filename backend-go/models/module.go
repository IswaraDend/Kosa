package models

// The 8 toggleable business feature modules a Super Admin can enable/disable
// per Project (see ProjectModule). Core screens (dashboard/summary, project
// members, permission management) are never gated — only these are.
const (
	ModuleWarehouse   = "warehouse"
	ModuleItem        = "item"
	ModuleProduct     = "product"
	ModuleProduction  = "production"
	ModuleTransaction = "transaction"
	ModuleCustomer    = "customer"
	ModuleInvoice     = "invoice"
	ModuleReport      = "report"
)

var AllModules = []string{
	ModuleWarehouse,
	ModuleItem,
	ModuleProduct,
	ModuleProduction,
	ModuleTransaction,
	ModuleCustomer,
	ModuleInvoice,
	ModuleReport,
}

// moduleDependencies maps a module to the other modules it requires to be
// meaningful (e.g. Item/Product stock is always tracked per Warehouse, so
// enabling either implies Warehouse). Enabling a module auto-enables its
// dependencies — see ExpandModules.
var moduleDependencies = map[string][]string{
	ModuleItem:        {ModuleWarehouse},
	ModuleProduct:     {ModuleWarehouse},
	ModuleProduction:  {ModuleProduct, ModuleItem, ModuleWarehouse},
	ModuleTransaction: {ModuleItem, ModuleWarehouse},
	ModuleInvoice:     {ModuleProduct, ModuleCustomer, ModuleWarehouse},
}

func IsValidModule(m string) bool {
	for _, v := range AllModules {
		if v == m {
			return true
		}
	}
	return false
}

// ExpandModules returns the given module set plus every transitive
// dependency, deduplicated. Invalid module codes are dropped silently — the
// caller is expected to have already surfaced a validation error for those
// if desired.
func ExpandModules(mods []string) []string {
	seen := map[string]bool{}

	var visit func(m string)
	visit = func(m string) {
		if !IsValidModule(m) || seen[m] {
			return
		}
		seen[m] = true
		for _, dep := range moduleDependencies[m] {
			visit(dep)
		}
	}

	for _, m := range mods {
		visit(m)
	}

	result := make([]string, 0, len(seen))
	for _, m := range AllModules {
		if seen[m] {
			result = append(result, m)
		}
	}
	return result
}
