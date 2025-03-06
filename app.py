from flask import Flask, request, render_template
from flask_login import login_required
from flask_wtf import FlaskForm
from wtforms import StringField, FloatField, SubmitField
from wtforms.validators import DataRequired
import logging
import os
import secrets

app = Flask(__name__)
app.config['SECRET_KEY'] = os.getenv('FLASK_SECRET_KEY', secrets.token_hex(32))  # Secure random key

# Logging setup
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# PaymentService class Example
class PaymentService:
    def process_payment(self, amount, card_details):
        try:
            if not amount > 0:
                raise ValueError("Invalid amount")
            return {"status": "success", "transaction_id": "12345"}
        except Exception as e:
            return {"status": "error", "message": str(e)}

payment_service = PaymentService()

# Form with CSRF protection
class PaymentForm(FlaskForm):
    amount = FloatField('Amount', validators=[DataRequired()])
    card_number = StringField('Card Number', validators=[DataRequired()])
    submit = SubmitField('Pay')

@app.route('/pay', methods=['GET', 'POST'])
@login_required
def pay():
    form = PaymentForm()
    if form.validate_on_submit():
        logger.info("Processing payment request")
        result = payment_service.process_payment(form.amount.data, form.card_number.data)
        if result['status'] == 'error':
            logger.error(f"Payment failed: {result['message']}")
            return render_template('payment.html', form=form, error=result['message'])
        return render_template('payment.html', form=form, success='Payment successful')
    return render_template('payment.html', form=form)

if __name__ == '__main__':
    app.run(debug=True)