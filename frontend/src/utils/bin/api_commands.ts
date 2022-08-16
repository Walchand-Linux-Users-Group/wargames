// // List of commands that require API calls

import { getRanks } from '../api';
import { getStats } from '../api';
import { submitKey } from '../api';

export const submit = async (args: string[]): Promise<string> => {

  if(args.length!=2){
    return ``
  }

  const username = args[0], key = args[1];

  const weather = await submitKey(username, key);
  return weather;
};

export const rank = async (args: string[]): Promise<string> => {

  if(args.length!=1){
    return `
    Please check the usage of rank command:
    
    +---------------+-----------------+--------------+
    |    Command    |      Usage      |   Example    |
    +===============+=================+==============+
    | Ranking Table | rank all        | rank all     |
    | User Stats    | rank {USERNAME} | rank suyashc |
    +---------------+-----------------+--------------+
    
    `
  }

  if (args[0]==="all") {
    const rankTable = await getRanks();
    return rankTable;
  }
  const userStats = await getStats(args[0]);
  return userStats;
};
