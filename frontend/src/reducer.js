import { IP } from './App'

export const ADD_EXTRACT_JOB = 'events/ADD_EXTRACT_JOB';
export const SUBMIT_ORGANIZED_JOBS = 'events/SUBMIT_ORGANIZED_JOBS'
export const SET_SELECTED_SUBREDDIT = 'events/SET_SELECTED_SUBREDDIT'
export const GET_SUBREDDITS = 'events/GET_SUBREDDITS'
export const RECEIVE_SUBREDDITS = 'events/RECEIVE_SUBREDDITS'

export default function reducer(state = [], action) {
  switch (action.type) {
    case ADD_EXTRACT_JOB:
      let newState = {}
      newState.extractionQueue = [];
      let found = false
      for (let job in state.extractionQueue) {
        let j = state.extractionQueue[job]
        if (j.subreddit === action.subreddit && j.month === action.month && j.year === action.year){
          found = true
        } else {
          newState.extractionQueue.push(j)
        }
      }

      if (!found) {
        newState.extractionQueue.push({
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
    
    case SET_SELECTED_SUBREDDIT:
      return {
        ...state,
        selectedSubreddit: action.subreddit
      }

    case GET_SUBREDDITS:
      let subs;
      fetch("http://" + IP + ":5000/api/subs")
        .then(results => {
          return results.json();
        })
        .then(data => {
            subs = data;
        });
      return {
        ...state, 
        subreddits: subs
      }
    case RECEIVE_SUBREDDITS:
      return {
        ...state, 
        subreddits: action.subreddits
      }
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

export function setSelectedSubreddit(subreddit) {
  return {
    type: SET_SELECTED_SUBREDDIT,
    subreddit: subreddit
  }
}

export function fetchSubreddits() {
  return {
    type: GET_SUBREDDITS
  }
}

export function submitOrganizedJobs(organized) {
  return {
    type: SUBMIT_ORGANIZED_JOBS,
    organized: organized
  };
}