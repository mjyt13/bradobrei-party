import type { UserDto } from './auth'
import type { SalonBriefDto } from './common'

export type BookingStatus =
  | 'PENDING'
  | 'CONFIRMED'
  | 'IN_PROGRESS'
  | 'COMPLETED'
  | 'CANCELLED'

export type PaymentStatus = 'PENDING' | 'SUCCESS' | 'FAILED' | 'REFUNDED'
export type SalonStatus = 'OPEN' | 'CLOSED'

export interface MaterialDto {
  id: number
  name: string
  unit: string
}

export interface ServiceMaterialDto {
  service_id: number
  material_id: number
  quantity_per_use: number
  material?: MaterialDto
}

export interface ServiceDto {
  id: number
  name: string
  description: string
  price: number
  duration_minutes: number
  created_at: string
  updated_at: string
  materials?: ServiceMaterialDto[]
  employees?: EmployeeProfileSummaryDto[]
}

export interface EmployeeProfileSummaryDto {
  id: number
  user_id: number
  specialization: string
  expected_salary: number
  work_schedule?: string
  user?: UserDto
  services?: ServiceDto[]
}

export interface SalonDto extends SalonBriefDto {
  /** Широта/долгота для карты в SPA (вычисляются на backend из PostGIS). */
  latitude?: number
  longitude?: number
  location?: string
  working_hours?: string
  status: SalonStatus
  max_staff: number
  base_hourly_rate: number
  created_at: string
  updated_at: string
  employees?: EmployeeProfileSummaryDto[]
}

export interface PaymentDto {
  id: number
  booking_id: number
  amount: number
  status: PaymentStatus
  external_transaction_id?: string
  created_at: string
  completed_at?: string
}

export interface BookingItemDto {
  id: number
  booking_id: number
  service_id: number
  quantity: number
  price_at_booking: number
  service?: ServiceDto
}

export interface BookingDto {
  id: number
  start_time: string
  duration_minutes: number
  status: BookingStatus
  total_price: number
  notes: string
  client_id: number
  master_id?: number
  salon_id: number
  created_at: string
  updated_at: string
  client?: UserDto
  master?: UserDto
  salon?: SalonDto
  items?: BookingItemDto[]
  payment?: PaymentDto
}

export interface GeocodeAddressResponseDto {
  latitude: number
  longitude: number
  formatted_address: string
  provider: string
}

export interface UpsertSalonRequestDto {
  name: string
  address: string
  location?: string
  working_hours?: string
  status: SalonStatus
  max_staff: number
  base_hourly_rate: number
}

export interface UpsertServiceRequestDto {
  name: string
  description: string
  price: number
  duration_minutes: number
}

export interface UpsertMaterialRequestDto {
  name: string
  unit: string
}

export interface CreatePaymentRequestDto {
  booking_id: number
  amount: number
  status: PaymentStatus
  external_transaction_id: string
}

export interface CreateBookingRequestDto {
  start_time: string
  salon_id: number
  master_id?: number
  service_ids: number[]
  notes: string
}
