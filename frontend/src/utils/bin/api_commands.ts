// // List of commands that require API calls

import { getLeaderboard } from '../api';
import { getStats } from '../api';
import { submitFlag } from '../api';

export const show = async (args: string[]): Promise<string> => {
  if (args.length != 1) {
    return `
  +---------------+--------------------------+
  |    Command    |          Usage           |
  +===============+==========================+
  | Submit        | submit {USERNAME}:{FLAG} |
  | Ranking Table | show all                 |
  | User Stats    | show {USERNAME}          |
  +---------------+--------------------------+
  
`;
  }

  var data;

  if(args[0]=="all"){
    data = await getLeaderboard();
  } else{
    data = await getStats(args[0]);
  }

  return data;
};

export const submit = async (args: string[]): Promise<string> => {
  console.log(args);
  if(args.length!=1){
    return `
  +---------------+--------------------------+
  |    Command    |          Usage           |
  +===============+==========================+
  | Submit        | submit {USERNAME}:{FLAG} |
  | Ranking Table | show all                 |
  | User Stats    | show {USERNAME}          |
  +---------------+--------------------------+
  
`;
  }

  var splitted = args[0].split(":", 2);
  
  if(splitted.length!=2){
    return `
  +---------------+--------------------------+
  |    Command    |          Usage           |
  +===============+==========================+
  | Submit        | submit {USERNAME}:{FLAG} |
  | Ranking Table | show all                 |
  | User Stats    | show {USERNAME}          |
  +---------------+--------------------------+
  
`;
  }

  const data = submitFlag(splitted[0],splitted[1]);

  return data;
};

