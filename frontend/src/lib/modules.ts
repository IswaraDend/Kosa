// Mirrors backend-go/models/module.go — keep the two in sync.

export type ModuleCode =
  | 'warehouse'
  | 'item'
  | 'product'
  | 'production'
  | 'transaction'
  | 'customer'
  | 'invoice'
  | 'report';

export const MODULES: { code: ModuleCode; label: string }[] = [
  { code: 'warehouse', label: 'Gudang' },
  { code: 'item', label: 'Item' },
  { code: 'product', label: 'Produk' },
  { code: 'production', label: 'Produksi' },
  { code: 'transaction', label: 'Transaksi' },
  { code: 'customer', label: 'Pelanggan' },
  { code: 'invoice', label: 'Invoice' },
  { code: 'report', label: 'Laporan' },
];

// The slice of MODULES this build actually puts on screen. Narrower than
// MODULES on purpose: for now the app only serves the Produk → Invoice path,
// so Item, Produksi, Transaksi, and Laporan are offered on no screen at all.
//
// Nothing was deleted to achieve that. The Go routes, handlers, and pages for
// those four are all still in place, and a project that still has one of them
// switched on keeps its ProjectModule rows and its member grants untouched —
// the feature is hidden, not removed. Putting a code back in this list is the
// whole of bringing it back.
//
// Gudang is gone from this list and from the schema it used to sit in: Invoice
// no longer carries a WarehouseID, and ProductStock is one balance per project.
// Pelanggan stays because Invoice still carries a CustomerID.
//
// The warehouses table itself is untouched — Item stock, Transaksi, and
// Produksi still count per warehouse, which is why 'warehouse' is still a
// module code and still a dependency of those three.
export const ACTIVE_MODULES: ModuleCode[] = ['product', 'customer', 'invoice'];

/** Whether this build exposes `code` at all, regardless of any project's own switch. */
export function isActiveModule(code: string): boolean {
  return (ACTIVE_MODULES as string[]).includes(code);
}

/** The modules a project may still have enabled but that this build hides. */
export function outOfScopeModules(enabled: string[]): ModuleCode[] {
  return MODULES.map((m) => m.code).filter((code) => enabled.includes(code) && !isActiveModule(code));
}

// A module implies its dependencies are enabled too (e.g. raw-material Item
// stock is always tracked per Gudang). Mirrors moduleDependencies in module.go.
//
// Product deliberately has no dependency: alone it is just a catalogue of
// finished goods. Invoice adds only Pelanggan, because finished-goods stock is
// counted per project rather than per warehouse. Produksi is the one that
// needs Gudang, since it consumes Item stock, which still is per warehouse.
export const MODULE_DEPENDENCIES: Record<ModuleCode, ModuleCode[]> = {
  warehouse: [],
  item: ['warehouse'],
  product: [],
  production: ['product', 'item', 'warehouse'],
  transaction: ['item', 'warehouse'],
  customer: [],
  invoice: ['product', 'customer'],
  report: [],
};

/** Returns the given module set plus every transitive dependency, deduped. */
export function expandModules(selected: string[]): string[] {
  const seen = new Set<string>();

  const visit = (m: string) => {
    if (seen.has(m) || !(m in MODULE_DEPENDENCIES)) return;
    seen.add(m);
    for (const dep of MODULE_DEPENDENCIES[m as ModuleCode]) {
      visit(dep);
    }
  };

  selected.forEach(visit);
  return MODULES.map((m) => m.code).filter((code) => seen.has(code));
}

/** Human label for a module code, falling back to the raw code if unknown. */
export function moduleLabel(code: string): string {
  return MODULES.find((m) => m.code === code)?.label ?? code;
}

/** Modules that cannot function without `m` — the inverse of MODULE_DEPENDENCIES. */
export function dependentsOf(m: ModuleCode): ModuleCode[] {
  return (Object.keys(MODULE_DEPENDENCIES) as ModuleCode[]).filter((code) =>
    MODULE_DEPENDENCIES[code].includes(m),
  );
}

/**
 * Removes `remove` from `current` along with everything that transitively
 * depends on it. Counterpart to expandModules — mirrors CollapseModules in
* module.go. Without this, unchecking a dependency (e.g. Gudang while Produksi
 * is on) would be undone immediately by the next expandModules call, making
 * the checkbox look stuck.
 */
export function collapseModules(current: string[], remove: ModuleCode): string[] {
  const dropped = new Set<string>();

  const visit = (m: ModuleCode) => {
    if (dropped.has(m)) return;
    dropped.add(m);
    dependentsOf(m).forEach(visit);
  };
  visit(remove);

  return MODULES.map((m) => m.code).filter((code) => current.includes(code) && !dropped.has(code));
}

// Pañca Kośa — the five layers a project's modules are grouped under,
// outer to inner. Ānandamaya (the innermost/"core" layer) has no
// toggleable module of its own — Ringkasan/Dashboard already isn't gated
// by any module, which is what makes it "always present" — so it's
// handled as a standalone item wherever this list is consumed, not a band.
export interface KosaLayer {
  code: 'anna' | 'prana' | 'mano' | 'vijnana';
  name: string;
  deva: string;
  gloss: string;
  colorVar: string;
  modules: ModuleCode[];
}

export const LAYERS: KosaLayer[] = [
  { code: 'anna', name: 'Annamaya', deva: 'अन्नमय', gloss: 'Fisik', colorVar: '--k-anna', modules: ['warehouse', 'item'] },
  { code: 'prana', name: 'Prāṇamaya', deva: 'प्राणमय', gloss: 'Napas', colorVar: '--k-prana', modules: ['transaction', 'production'] },
  { code: 'mano', name: 'Manomaya', deva: 'मनोमय', gloss: 'Pikiran', colorVar: '--k-mano', modules: ['product', 'customer'] },
  { code: 'vijnana', name: 'Vijñānamaya', deva: 'विज्ञानमय', gloss: 'Kebijaksanaan', colorVar: '--k-vijnana', modules: ['invoice', 'report'] },
];

// Shared sidebar-grouping types/helper — used by both DashboardLayout (Admin)
// and MemberLayout so a project's enabled modules render as the same
// Pañca Kośa bands regardless of role.
export interface NavLeaf {
  path: string;
  label: string;
  icon: import('react').ReactNode;
}

export type SidebarBlock = { kind: 'item'; item: NavLeaf } | { kind: 'group'; layerIdx: number; items: NavLeaf[] };

/**
 * Builds the Pañca Kośa sidebar bands.
 *
 * A module may contribute more than one menu entry — Produk brings Kategori
 * with it — so a value here is either one leaf or several. They land next to
 * each other inside their module's band, which is where a reader expects them.
 */
export function groupByLayer(
  enabledModules: string[],
  nav: Partial<Record<ModuleCode, NavLeaf | NavLeaf[]>>,
): SidebarBlock[] {
  const blocks: SidebarBlock[] = [];
  LAYERS.forEach((layer, layerIdx) => {
    const items = layer.modules
      // isActiveModule is the second gate on purpose: a project created before
      // the scope narrowed may still carry item/production/transaction/report
      // in its ProjectModule rows, and those rows are deliberately left alone.
      // Filtering here keeps both the Admin and the Member sidebar correct
      // whatever the database still says.
      .filter((m) => enabledModules.includes(m) && isActiveModule(m))
      .flatMap((m) => {
        const entry = nav[m];
        if (!entry) return [];
        return Array.isArray(entry) ? entry : [entry];
      });
    if (items.length > 0) blocks.push({ kind: 'group', layerIdx, items });
  });
  return blocks;
}
