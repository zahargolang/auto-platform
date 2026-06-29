import { apiFetch } from '../lib/apiClient'
import type { Listing, ListingFilters, ListingFormValues, ListingsResponse } from '../types'

function buildQuery(filters: ListingFilters): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== null && value !== '') {
      params.set(key, String(value))
    }
  }
  const qs = params.toString()
  return qs ? `?${qs}` : ''
}

export function getListings(filters: ListingFilters = {}): Promise<ListingsResponse> {
  return apiFetch<ListingsResponse>(`/listings${buildQuery(filters)}`, { auth: false })
}

export function getListing(id: string): Promise<Listing> {
  return apiFetch<Listing>(`/listings/${id}`, { auth: false })
}

export function createListing(values: ListingFormValues): Promise<Listing> {
  return apiFetch<Listing>('/listings/mine', { method: 'POST', body: values })
}

export function updateListing(
  id: string,
  values: Partial<ListingFormValues & { status: string }>,
): Promise<Listing> {
  return apiFetch<Listing>(`/listings/mine/${id}`, { method: 'PATCH', body: values })
}

export function deleteListing(id: string): Promise<void> {
  return apiFetch<void>(`/listings/mine/${id}`, { method: 'DELETE' })
}
