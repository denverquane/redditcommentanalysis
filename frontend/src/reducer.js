export const ADD_EXTRACT_JOB = 'events/ADD_EXTRACT_JOB';
export const ADD_EXTRACT_JOB_SUCCESS = 'events/ADD_EXTRACT_JOB_SUCCESS';
export const ADD_EXTRACT_JOB_FAIL = 'events/ADD_EXTRACT_JOB_FAIL';

export default function reducer(state = { extractionJobs: [] }, action) {
  switch (action.type) {
    case ADD_EXTRACT_JOB:
      return { ...state };
    case ADD_EXTRACT_JOB_SUCCESS:
      return { ...state, extractionJobs: [action.payload.job] };
    case ADD_EXTRACT_JOB_FAIL:
      return {
        ...state,
        error: 'Error while fetching events'
      };
    default:
      return state;
  }
}

export function addExtractionJob(job) {
  return {
    type: ADD_EXTRACT_JOB,
    payload: {
      job: job
    }
  };
}