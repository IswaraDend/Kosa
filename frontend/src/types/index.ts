export interface AuthUser {
  id: number;
  name: string;
  email: string;
  is_super_admin: boolean;
  is_admin: boolean;
  is_member: boolean;
}

export interface Project {
  id: number;
  name: string;
  code: string;
  description: string;
  status: 'aktif' | 'nonaktif';
  created_by: number;
  created_at: string;
  admin_name: string;
  admin_count: number;
  member_count: number;
  warehouse_count: number;
  modules: string[];
}

export interface UserListItem {
  id: number;
  name: string;
  email: string;
  is_super_admin: boolean;
  user_role_id: number;
  role: 'admin' | 'member';
  project_id: number;
  project_name: string;
  created_at: string;
}

export interface Role {
  id: number;
  project_id: number;
  name: string;
  description: string;
}

export interface PermissionDef {
  id: number;
  code: string;
  name: string;
  module: string;
}

export interface MemberPermission {
  id: number;
  user_role_id: number;
  permission_id: number;
  granted_by: number;
  granted_at: string;
  Permission?: PermissionDef;
}

export interface Warehouse {
  id: number;
  project_id: number;
  name: string;
  code: string;
  address: string;
  created_at: string;
}

export interface Item {
  id: number;
  project_id: number;
  sku: string;
  name: string;
  unit: string;
  average_cost: number;
  created_at: string;
}

export type TransactionType = 'in' | 'out' | 'transfer';

export interface TransactionItemLine {
  item_id: number;
  quantity: number;
  unit_cost?: number;
}

export interface TransactionRecord {
  id: number;
  project_id: number;
  type: TransactionType;
  source_warehouse_id: number | null;
  dest_warehouse_id: number | null;
  note: string;
  performed_by: number;
  created_at: string;
  items: (TransactionItemLine & { id: number; item?: Item })[];
  source_warehouse?: Warehouse;
  dest_warehouse?: Warehouse;
}

export interface SummaryData {
  total_projects: number;
  active_projects: number;
  total_warehouses: number;
  total_admins: number;
  total_members: number;
  total_invoices: number;
  total_revenue: number;
  total_margin: number;
}

export interface StockSummaryRow {
  item_id: number;
  item_name: string;
  unit: string;
  warehouse_id: number;
  warehouse_name: string;
  quantity: number;
}

export interface TransactionReportRow {
  period: string;
  type: TransactionType;
  total_qty: number;
}

export interface SalesSummaryRow {
  period: string;
  total_qty: number;
  subtotal: number;
  total_hpp: number;
}

export interface TopProductRow {
  product_id: number;
  product_name: string;
  total_qty_sold: number;
  total_revenue: number;
}

export interface Product {
  id: number;
  project_id: number;
  sku: string;
  name: string;
  unit: string;
  average_cost: number;
  default_price: number;
  created_at: string;
}

export interface ProductRecipeLine {
  id: number;
  project_id: number;
  product_id: number;
  item_id: number;
  quantity_per_unit: number;
  created_at: string;
  item?: Item;
}

export interface ProductStockRow {
  id: number;
  project_id: number;
  warehouse_id: number;
  product_id: number;
  quantity: number;
  updated_at: string;
  warehouse?: Warehouse;
}

export interface Customer {
  id: number;
  project_id: number;
  name: string;
  phone: string;
  email: string;
  address: string;
  created_at: string;
}

export type InvoiceStatus = 'unpaid' | 'paid' | 'cancelled';

export interface InvoiceItemLine {
  id: number;
  invoice_id: number;
  product_id: number;
  quantity: number;
  unit_price: number;
  unit_cogs: number;
  product?: Product;
}

export interface InvoiceRecord {
  id: number;
  project_id: number;
  invoice_number: string;
  customer_id: number;
  warehouse_id: number;
  status: InvoiceStatus;
  subtotal: number;
  total_hpp: number;
  note: string;
  performed_by: number;
  created_at: string;
  customer?: Customer;
  warehouse?: Warehouse;
  items: InvoiceItemLine[];
}

export interface ProductionRecord {
  id: number;
  project_id: number;
  warehouse_id: number;
  product_id: number;
  quantity: number;
  hpp_per_unit: number;
  hpp_total: number;
  note: string;
  transaction_id: number | null;
  performed_by: number;
  created_at: string;
  product?: Product;
  warehouse?: Warehouse;
  transaction?: TransactionRecord;
}
