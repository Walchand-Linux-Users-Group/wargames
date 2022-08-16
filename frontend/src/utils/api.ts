import axios from 'axios';
import config from '../../config.json';

export const getRanks = async () => {
  const { data } = await axios.get(
    `https://api.wargames.wcewlug.org/leaderboard`,
  );
  return data;
};

export const getStats = async (username) => {
  const { data } = await axios.get(
    `https://api.wargames.wcewlug.org/stats`,
  );
  return data;
};

export const submitKey = async (username, key) => {
  const { data } = await axios.get(
    `https://api.wargames.wcewlug.org/verify`,
  );
  return data;
};
