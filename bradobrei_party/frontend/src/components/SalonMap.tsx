import { useEffect, useRef } from 'react'
import type { SalonDto } from '../types/dto/entities'

type MapProvider = 'yandex' | 'google' | 'none'

declare global {
  interface Window {
    ymaps?: {
      ready: (cb: () => void) => void
      Map: new (el: HTMLElement, state: { center: number[]; zoom: number }) => {
        geoObjects: { add: (obj: unknown) => void; removeAll: () => void }
        destroy: () => void
      }
      Placemark: new (geometry: number[], props?: Record<string, string>) => unknown
    }
    google?: {
      maps: {
        Map: new (el: HTMLElement, opts: { center: { lat: number; lng: number }; zoom: number }) => {
          setCenter: (c: { lat: number; lng: number }) => void
          setZoom: (z: number) => void
        }
        Marker: new (opts: { position: { lat: number; lng: number }; map: unknown; title?: string }) => void
        LatLngBounds: new () => { extend: (p: { lat: number; lng: number }) => void }
      }
    }
  }
}

async function loadScriptOnce(id: string, src: string, isReady: () => boolean): Promise<void> {
  if (isReady()) {
    return
  }
  const existing = document.getElementById(id) as HTMLScriptElement | null
  if (existing?.src === src) {
    const deadline = Date.now() + 8000
    while (Date.now() < deadline) {
      if (isReady()) {
        return
      }
      await new Promise((r) => setTimeout(r, 50))
    }
    throw new Error('таймаут загрузки скрипта карты')
  }
  await new Promise<void>((resolve, reject) => {
    const script = document.createElement('script')
    script.id = id
    script.async = true
    script.src = src
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Не удалось загрузить скрипт карты'))
    document.head.appendChild(script)
  })
}

function salonsWithCoords(salons: SalonDto[]) {
  return salons.filter(
    (s): s is SalonDto & { latitude: number; longitude: number } =>
      typeof s.latitude === 'number' && typeof s.longitude === 'number',
  )
}

export function SalonMap({ salons }: { salons: SalonDto[] }) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<{ destroy?: () => void } | null>(null)
  const markersRef = useRef<unknown[]>([])

  const rawProvider = import.meta.env.VITE_MAP_PROVIDER?.toLowerCase() ?? 'none'
  const provider: MapProvider = rawProvider === 'yandex' || rawProvider === 'google' ? rawProvider : 'none'
  const yandexKey = import.meta.env.VITE_YANDEX_MAPS_JS_API_KEY ?? ''
  const googleKey = import.meta.env.VITE_GOOGLE_MAPS_JS_API_KEY ?? ''

  useEffect(() => {
    const el = containerRef.current
    if (!el) {
      return
    }

    mapRef.current?.destroy?.()
    mapRef.current = null
    markersRef.current = []
    el.innerHTML = ''

    const withCoords = salonsWithCoords(salons)
    if (withCoords.length === 0) {
      return
    }

    let cancelled = false

    void (async () => {
      if (provider === 'none') {
        return
      }

      if (provider === 'yandex') {
        if (!yandexKey) {
          return
        }
        try {
          await loadScriptOnce(
            'bradobrei-yandex-maps',
            `https://api-maps.yandex.ru/2.1/?apikey=${encodeURIComponent(yandexKey)}&lang=ru_RU`,
            () => Boolean(window.ymaps),
          )
        } catch {
          return
        }
        if (cancelled || !window.ymaps || !containerRef.current) {
          return
        }
        const ymaps = window.ymaps
        ymaps.ready(() => {
          if (cancelled || !containerRef.current) {
            return
          }
          const center: [number, number] = [withCoords[0].latitude, withCoords[0].longitude]
          const map = new ymaps.Map(containerRef.current, {
            center,
            zoom: withCoords.length === 1 ? 14 : 10,
          })
          mapRef.current = map
          withCoords.forEach((s) => {
            const placemark = new ymaps.Placemark([s.latitude, s.longitude], {
              balloonContent: `${s.name}<br/>${s.address}`,
            })
            map.geoObjects.add(placemark)
          })
        })
        return
      }

      if (provider === 'google') {
        if (!googleKey) {
          return
        }
        try {
          await loadScriptOnce(
            'bradobrei-google-maps',
            `https://maps.googleapis.com/maps/api/js?key=${encodeURIComponent(googleKey)}`,
            () => Boolean(window.google?.maps),
          )
        } catch {
          return
        }
        if (cancelled || !window.google?.maps || !containerRef.current) {
          return
        }
        const g = window.google.maps
        const center = { lat: withCoords[0].latitude, lng: withCoords[0].longitude }
        const map = new g.Map(containerRef.current, {
          center,
          zoom: withCoords.length === 1 ? 14 : 10,
        })
        mapRef.current = map as unknown as { destroy?: () => void }
        withCoords.forEach((s) => {
          const marker = new g.Marker({
            position: { lat: s.latitude, lng: s.longitude },
            map,
            title: s.name,
          })
          markersRef.current.push(marker)
        })
      }
    })()

    return () => {
      cancelled = true
      mapRef.current?.destroy?.()
      mapRef.current = null
      markersRef.current = []
    }
  }, [salons, provider, yandexKey, googleKey])

  const withCoords = salonsWithCoords(salons)

  if (provider === 'none') {
    return (
      <div className="salon-map-panel salon-map-hint">
        <p className="eyebrow">Карта</p>
        <p>Укажите в <code>.env</code> фронтенда <code>VITE_MAP_PROVIDER=yandex</code> или <code>google</code> и ключ JS API для отображения карты (только для UI; геокодер остаётся на сервере).</p>
      </div>
    )
  }

  if ((provider === 'yandex' && !yandexKey) || (provider === 'google' && !googleKey)) {
    return (
      <div className="salon-map-panel salon-map-hint">
        <p className="eyebrow">Карта</p>
        <p>Добавьте ключ картографического UI: {provider === 'yandex' ? 'VITE_YANDEX_MAPS_JS_API_KEY' : 'VITE_GOOGLE_MAPS_JS_API_KEY'}.</p>
      </div>
    )
  }

  if (withCoords.length === 0) {
    return (
      <div className="salon-map-panel salon-map-hint">
        <p className="eyebrow">Карта</p>
        <p>Нет салонов с координатами. Укажите адрес и при необходимости включите серверный геокодер или задайте координаты вручную.</p>
      </div>
    )
  }

  return (
    <div className="salon-map-panel">
      <div className="salon-map-head">
        <p className="eyebrow">Карта филиалов</p>
        <p className="section-description">
          Тайлы и SDK загружаются в браузер (ограниченный ключ). Координаты для маркеров приходят из API после валидации на сервере.
        </p>
      </div>
      <div ref={containerRef} className="salon-map-canvas" role="presentation" />
    </div>
  )
}
