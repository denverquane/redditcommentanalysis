import React from 'react';
import ReactDOM from 'react-dom';
import {Provider} from 'react-redux'
import {createStore, applyMiddleware} from 'redux';
import axios from 'axios';
import axiosMiddleware from 'redux-axios-middleware';
import './index.css';
import App from './App';
import * as serviceWorker from './serviceWorker';
import reducer from './reducer'

export const IP = "localhost";

const client = axios.create({ //all axios can be used, shown in axios documentation
  baseURL:'http://' + IP + ':5000/api',
  responseType: 'json'
});

const store = createStore(reducer,
  applyMiddleware(
    axiosMiddleware(client)
  ))

ReactDOM.render(
<Provider store={store}>
    <App />
  </Provider>, document.getElementById('root')
);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: http://bit.ly/CRA-PWA
serviceWorker.unregister();
