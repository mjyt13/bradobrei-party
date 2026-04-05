import { apiRequest } from '../client'
import type {
  EmployeeProfileSummaryDto,
  GeocodeAddressResponseDto,
  SalonDto,
  UpsertSalonRequestDto,
} from '../../types/dto/entities'

export const salonService = {
  getAll() {
    return apiRequest<SalonDto[]>('/salons')
  },
  getById(id: number) {
    return apiRequest<SalonDto>(`/salons/${id}`)
  },
  getMasters(id: number) {
    return apiRequest<EmployeeProfileSummaryDto[]>(`/salons/${id}/masters`)
  },
  create(payload: UpsertSalonRequestDto) {
    return apiRequest<SalonDto>('/salons', {
      method: 'POST',
      body: payload,
    })
  },
  update(id: number, payload: UpsertSalonRequestDto) {
    return apiRequest<SalonDto>(`/salons/${id}`, {
      method: 'PUT',
      body: payload,
    })
  },
  remove(id: number) {
    return apiRequest<{ message: string }>(`/salons/${id}`, {
      method: 'DELETE',
    })
  },
  /** Серверный геокодер (секрет провайдера только на backend). */
  geocodeAddress(address: string) {
    return apiRequest<GeocodeAddressResponseDto>('/salons/geocode', {
      method: 'POST',
      body: { address },
    })
  },
}
