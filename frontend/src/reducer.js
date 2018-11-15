export const ADD_EXTRACT_JOB = 'events/ADD_EXTRACT_JOB';

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
    default:
      return state;
  }
}

export function addExtractionJob(subreddit, month, year) {
  return {
    type: ADD_EXTRACT_JOB,
    subreddit: subreddit,
    month: month,
    year: year
  };
}