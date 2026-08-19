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
// meaningful (e.g. raw-material Item stock is always tracked per Warehouse,
// so enabling Item implies Warehouse). Enabling a module auto-enables its
// dependencies — see ExpandModules.
//
// Product deliberately has NO dependency: on its own it is a plain catalogue
// of finished goods (SKU/name/unit/price). Only the features that actually
// move product quantity around — Production and Invoice — pull in Warehouse,
// and Production additionally pulls in Item for its recipe (BOM).
var moduleDependencies = map[string][]string{
	ModuleItem:        {ModuleWarehouse},
	ModuleProduction:  {ModuleProduct, ModuleItem, ModuleWarehouse},
	ModuleTransaction: {ModuleItem, ModuleWarehouse},
	ModuleInvoice:     {ModuleProduct, ModuleCustomer, ModuleWarehouse},
}

// DependentsOf returns the modules that cannot function without m — the
// inverse of moduleDependencies. Turning a module off must also turn these
// off, otherwise the stored set would claim e.g. Invoice is enabled while
// its required Warehouse is not. See CollapseModules.
func DependentsOf(m string) []string {
	var dependents []string
	for module, deps := range moduleDependencies {
		for _, d := range deps {
			if d == m {
				dependents = append(dependents, module)
				break
			}
		}
	}
	return dependents
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

// CollapseModules removes `remove` from `current` along with every module
// that transitively depends on it, returning the result in AllModules order.
// This is the counterpart to ExpandModules: expanding pulls dependencies in,
// collapsing pushes dependents out.
func CollapseModules(current []string, remove string) []string {
	dropped := map[string]bool{}

	var visit func(m string)
	visit = func(m string) {
		if dropped[m] {
			return
		}
		dropped[m] = true
		for _, dep := range DependentsOf(m) {
			visit(dep)
		}
	}
	visit(remove)

	result := make([]string, 0, len(current))
	for _, m := range AllModules {
		for _, c := range current {
			if c == m && !dropped[m] {
				result = append(result, m)
				break
			}
		}
	}
	return result
}
