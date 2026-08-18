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

// A module implies its dependencies are enabled too (e.g. Item/Product stock
// is always tracked per Gudang). Mirrors moduleDependencies in module.go.
export const MODULE_DEPENDENCIES: Record<ModuleCode, ModuleCode[]> = {
  warehouse: [],
  item: ['warehouse'],
  product: ['warehouse'],
  production: ['product', 'item', 'warehouse'],
  transaction: ['item', 'warehouse'],
  customer: [],
  invoice: ['product', 'customer', 'warehouse'],
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

export function groupByLayer(enabledModules: string[], nav: Partial<Record<ModuleCode, NavLeaf>>): SidebarBlock[] {
  const blocks: SidebarBlock[] = [];
  LAYERS.forEach((layer, layerIdx) => {
    const items = layer.modules
      .filter((m) => enabledModules.includes(m))
      .map((m) => nav[m])
      .filter((n): n is NavLeaf => !!n);
    if (items.length > 0) blocks.push({ kind: 'group', layerIdx, items });
  });
  return blocks;
}
