import api from '$lib/api';

export async function load() {
  try {
    const res = await api.get('/admin/badges');
    return {
      badges: res.data.data || []
    };
  } catch (error) {
    console.error('Error loading badges:', error);
    return {
      badges: []
    };
  }
}
