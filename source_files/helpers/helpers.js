/**
 * Formats a date object into YYYY-MM-DD string.
 * @param {Date} date - The date to format.
 * @returns {string} The formatted date string.
 */
 function formatDate(date) {
    if (!date) {
      return new Date().toISOString().split('T')[0];
    }
    return date.toISOString().split('T')[0];
  }
  
  /**
   * Validates an email address.
   * @param {string} email - The email to validate.
   * @returns {boolean} True if the email is valid, false otherwise.
   */
  function validateEmail(email) {
    const re = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
  }
  
  module.exports = {
    formatDate,
    validateEmail
  };
  