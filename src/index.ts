// index.ts
import { fetchContributions } from './githubApi';
import { calculateStreak } from './streakTracker';

export async function getStreak(username: string): Promise<number> {
  const contributions = await fetchContributions(username);
  return calculateStreak(contributions);
}

