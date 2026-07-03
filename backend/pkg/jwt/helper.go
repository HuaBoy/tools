package jwt

// ExpireHours 返回过期时间（小时）
func (m *Manager) ExpireHours() int {
	return m.expireHours
}
