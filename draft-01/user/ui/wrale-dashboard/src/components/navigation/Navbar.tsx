import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { 
  HomeIcon, 
  ServerIcon, 
  MapPinIcon, 
  ChartBarIcon, 
  CogIcon 
} from '@heroicons/react/24/outline'

const navigation = [
  { name: 'Dashboard', href: '/', icon: HomeIcon },
  { name: 'Devices', href: '/devices', icon: ServerIcon },
  { name: 'Physical Map', href: '/map', icon: MapPinIcon },
  { name: 'Analytics', href: '/analytics', icon: ChartBarIcon },
  { name: 'Settings', href: '/settings', icon: CogIcon },
]

export function Navbar() {
  const pathname = usePathname()

  return (
    <nav className="bg-wrale-primary h-screen w-64 fixed left-0 top-0 text-white p-4">
      <div className="mb-8">
        <h1 className="text-xl font-bold">Wrale Fleet</h1>
        <p className="text-sm text-gray-300">Physical-First Management</p>
      </div>

      <div className="space-y-2">
        {navigation.map((item) => {
          const isActive = pathname === item.href
          const Icon = item.icon
          
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`flex items-center px-4 py-2 rounded-lg transition-colors ${
                isActive 
                  ? 'bg-white/10 text-white' 
                  : 'text-gray-300 hover:bg-white/5 hover:text-white'
              }`}
            >
              <Icon className="w-5 h-5 mr-3" />
              {item.name}
            </Link>
          )
        })}
      </div>
    </nav>
  )
}