import { apiFetch } from '../lib/apiClient'
import type { UploadURLResponse } from '../types'

function getUploadURL(filename: string): Promise<UploadURLResponse> {
  return apiFetch<UploadURLResponse>('/storage/mine/upload-url', {
    method: 'POST',
    body: { filename },
  })
}

// Загружает файл напрямую в S3 по presigned URL — байты файла никогда не
// проходят через наш backend, только сам URL (см. storage-service).
export async function uploadFile(file: File): Promise<string> {
  const { upload_url, public_url } = await getUploadURL(file.name)

  const res = await fetch(upload_url, {
    method: 'PUT',
    body: file,
    headers: { 'Content-Type': file.type },
  })
  if (!res.ok) {
    throw new Error(`Не удалось загрузить файл (${res.status})`)
  }

  return public_url
}
