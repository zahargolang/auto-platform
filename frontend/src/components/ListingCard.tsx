import { Link } from 'react-router-dom'
import type { Listing } from '../types'
import { formatPrice } from '../lib/constants'

export function ListingCard({ listing }: { listing: Listing }) {
  return (
    <Link
      to={`/listings/${listing.id}`}
      className="block rounded-lg border border-gray-200 bg-white p-4 transition hover:border-blue-400 hover:shadow-md"
    >
      <h3 className="text-lg font-semibold">
        {listing.make} {listing.model}, {listing.year}
      </h3>
      <p className="mt-1 text-xl font-bold text-blue-600">{formatPrice(listing.price)}</p>
      <p className="mt-2 text-sm text-gray-600">
        {listing.mileage.toLocaleString('ru-RU')} км · {listing.city}
      </p>
      {listing.status !== 'active' && (
        <span className="mt-2 inline-block rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-500">
          {listing.status === 'sold' ? 'продано' : 'неактивно'}
        </span>
      )}
    </Link>
  )
}
