import type { UserDto } from './dto/auth'

export interface AppShellOutletContext {
  currentUser: UserDto | null
}
