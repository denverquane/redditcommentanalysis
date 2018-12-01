import { IP } from "./index";

export const ADD_EXTRACT_JOB = "events/ADD_EXTRACT_JOB";
export const SET_SELECTED_SUBREDDIT = "events/SET_SELECTED_SUBREDDIT";
export const GET_SUBREDDITS = "events/GET_SUBREDDITS";
export const GET_SUBREDDITS_SUCCESS = "events/GET_SUBREDDITS_SUCCESS";

export const POST_EXTRACT_SUBREDDITS = "events/POST_EXTRACT_SUBREDDITS";
export const POST_EXTRACT_SUBREDDITS_SUCCESS =
  "events/POST_EXTRACT_SUBREDDITS_SUCCESS";

export default function reducer(state = [], action) {
  switch (action.type) {
    case ADD_EXTRACT_JOB:
      let newState = JSON.parse(JSON.stringify(state));
      newState.extractionQueue = [];
      let found = false;
      for (let job in state.extractionQueue) {
        let j = state.extractionQueue[job];
        if (
          j.subreddit === action.subreddit &&
          j.month === action.month &&
          j.year === action.year
        ) {
          found = true;
        } else {
          newState.extractionQueue.push(j);
        }
      }

      if (!found) {
        newState.extractionQueue.push({
          subreddit: action.subreddit,
          month: action.month,
          year: action.year
        });
      }
      return newState;

    case SET_SELECTED_SUBREDDIT:
      return {
        ...state,
        selectedSubreddit: action.subreddit
      };

    case GET_SUBREDDITS:
      return {
        ...state,
        loadingSubs: true
      };
    case GET_SUBREDDITS_SUCCESS:
      return {
        ...state,
        loadingSubs: false,
        subreddits: action.payload.data
      };
    case POST_EXTRACT_SUBREDDITS: 
      return {
        ...state,
        postingExtract: true
      }

    case POST_EXTRACT_SUBREDDITS_SUCCESS: 
      return {
        ...state,
        postingExtract: false,
        extractionQueue: []
      }

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

export function setSelectedSubreddit(subreddit) {
  return {
    type: SET_SELECTED_SUBREDDIT,
    subreddit: subreddit
  };
}

export function fetchSubreddits() {
  return {
    type: GET_SUBREDDITS,
    payload: {
      request: {
        url: "/subs"
      }
    }
  };
}

export function postExtractSubreddits(subs, month, year) {
  return {
    type: POST_EXTRACT_SUBREDDITS,
    payload: {
      request: {
        url: `/extractSubs/${month}/${year}`,
        method: "POST",
        data: JSON.stringify(subs) 
      }
    }
  };
}
