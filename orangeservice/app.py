from flask import Flask, render_template, flash, redirect, request
import Orange
app = Flask(__name__)
data = Orange.data.Table('../data/classification.csv')
learner = Orange.classification.CN2Learner()
classifier = learner(data)

c_values = data.domain.class_var.values
for d in data[12:24]:
    c = classifier(d)
    print("Predicted {}, originally {}".format(c_values[int(classifier(d)[0])],
                                     d.get_class()))

res = Orange.evaluation.CrossValidation(data, [learner], k=5)
print("Accuracy:", Orange.evaluation.scoring.CA(res))

@app.route('/', methods=['GET', 'POST'])
def prediction():
    if (request.method == 'POST'):
        jsonFile = request.files['file']
        # This is where we'd actually process the file, for now
        # we're just printing the lines of the file
        
        for line in jsonFile:
            print(line)
        return '''
        <!doctype html>
        <title>Predicted Subreddit</title>
        <h2>This is where I'd put the subreddit prediction<br/>IF I HAD ONE.
        </h2>
        '''
    return render_template('prediction.html')
if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
