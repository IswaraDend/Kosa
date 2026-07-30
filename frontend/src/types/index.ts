export interface AuthUser {
  id: number;
  name: string;
  email: string;
  is_super_admin: boolean;
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
  created_at: string;
}

export type TransactionType = 'in' | 'out' | 'transfer';

export interface TransactionItemLine {
  item_id: number;
  quantity: number;
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
