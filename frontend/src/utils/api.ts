import axios from 'axios';
import config from '../../config.json';

export const getLeaderboard = async () => {
  const data = await axios({
    method: 'get',
    url: "https://wargames.wcewlug.org:8888/leaderboard",
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Content-Type': 'application/json',},
  });

  return JSON.stringify(data['data'], null, 2);
};

export const getStats = async (username) => {
  const data = await axios({
    method: 'post',
    url: "https://wargames.wcewlug.org:8888/stats",
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Content-Type': 'application/json',
    }, 
    data: {
      username: username,
    }
  });

  return JSON.stringify(data['data'], null, 2);
};

export const submitFlag = async (username, flag) => {

  const data = await axios({
    method: 'post',
    url: "https://wargames.wcewlug.org:8888/verify",
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Content-Type': 'application/json',
    }, 
    data: {
      username: username,
      flag: flag
    }
  });

  return JSON.stringify(data['data'], null, 2);
};
