// Web app that marks a blessing as sent in the blessing sheet (first tab,
// columns: blessing, Sent; row 1 is the header).
//
// Setup: in the sheet, Extensions > Apps Script, paste this file, then
// Deploy > New deployment > Web app, Execute as: Me, Who has access: Anyone.
// Store the /exec URL in BLESSING_UPDATE_LINK. Anyone holding that URL can
// tick rows, so keep it secret.

function doPost(e) {
  const blessing = JSON.parse(e.postData.contents).blessing;
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheets()[0];
  const rows = sheet.getDataRange().getValues();

  for (let i = 1; i < rows.length; i++) {
    const sent = String(rows[i][1]).toUpperCase() === 'TRUE';
    if (String(rows[i][0]).trim() === blessing && !sent) {
      sheet.getRange(i + 1, 2).setValue(true);
      return ContentService.createTextOutput('ok');
    }
  }

  return ContentService.createTextOutput('not found');
}
