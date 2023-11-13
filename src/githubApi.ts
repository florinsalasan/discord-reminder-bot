import { Octokit } from '@octokit/core';
import * as dotenv from 'dotenv';

dotenv.config();

interface GitHubStreakTrackerConfig {
  token: string;
}

const octokit = new Octokit({
  auth: `${process.env.GITHUB_TOKEN}`, // Include 'token' here
  userAgent: 'GitHub-Streak-Tracker', // Add a custom user agent if desired
});

export async function fetchContributions(config: GitHubStreakTrackerConfig, page: number = 1): Promise<number[]> {
  const { token } = config;

  try {
    // Request to get the authenticated user's information
    const userResponse = await octokit.request<any>('GET /user');

    // Extract the username from the authenticated user's information
    const username = userResponse.data.login;

    // Request to get the events for the authenticated user
    const response = await octokit.request<any>('GET /users/{username}/events', {
      username,
      per_page: 100, // Adjust per_page as needed (max 100)
      page,
    });

    // Extract the contributions for each day from all types of events
    const contributions = response.data.map((event: any) => {
      const date = new Date(event.created_at).toLocaleDateString();
      // Adjust this part based on the properties available in different event types
      return { date, count: event.payload.size || 1 }; // Use a property that represents the contribution count
    });

    return contributions;
  } catch (error: unknown) {
    if (error instanceof Error) {
      // Handle errors
      console.error('Error fetching contributions:', error.message);
      throw error;
    } else {
      console.error('Unknown error:', error);
      throw new Error('Unknown error occurred.');
    }
  }
}
