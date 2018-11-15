import { IP } from './App'

export const ADD_EXTRACT_JOB = 'events/ADD_EXTRACT_JOB';
export const SUBMIT_ORGANIZED_JOBS = 'events/SUBMIT_ORGANIZED_JOBS'

export default function reducer(state = [], action) {
  switch (action.type) {
    case ADD_EXTRACT_JOB:
      let newState = []
      let found = false
      for (let job in state) {
        let j = state[job]
        if (j.subreddit === action.subreddit && j.month === action.month && j.year === action.year){
          found = true
        } else {
          newState.push(j)
        }
      }
      if (!found) {
        newState.push({
          subreddit: action.subreddit,
          month: action.month,
          year: action.year
        })
      }
      return newState

    case SUBMIT_ORGANIZED_JOBS: 
      for (let yrIdx in action.organized){
        for (let moIdx in action.organized[yrIdx]){
          let subs = action.organized[yrIdx][moIdx]

          extractSubreddits(subs, moIdx, yrIdx)
        }
      }
      return []
    default:
      return state;
  }
}

function extractSubreddits(subs, month, year) {
  fetch(
    "http://" + IP + ":5000/api/extractSubs/" + month + "/" + year,
    {
      method: "post",
      body: JSON.stringify(subs)
    }
  )
    .then(results => {
      return results;
    })
    .then(data => {
      console.log(data);
    });
}

export function addExtractionJob(subreddit, month, year) {
  return {
    type: ADD_EXTRACT_JOB,
    subreddit: subreddit,
    month: month,
    year: year
  };
}

export function submitOrganizedJobs(organized) {
  return {
    type: SUBMIT_ORGANIZED_JOBS,
    organized: organized
  };
}